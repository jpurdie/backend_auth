package app

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/jpurdie/authapi"
	authMw "github.com/jpurdie/authapi/pkg/utl/middleware/auth"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrRoleNotFound = authapi.ErrorResp{Error: authapi.Error{CodeInt: http.StatusUnprocessableEntity, Message: "Invalid role"}}
	ErrUserNotFound = authapi.ErrorResp{Error: authapi.Error{CodeInt: http.StatusNotFound, Message: "User not found"}}
	ErrInternal     = authapi.ErrorResp{Error: authapi.Error{CodeInt: http.StatusConflict, Message: "There was a problem"}}
)

type UserStore interface {
	List(o *authapi.Organization) ([]authapi.OrganizationUser, error)
	ListRoles() ([]authapi.Role, error)
	Update(ou authapi.OrganizationUser) error
}

type UserResource struct {
	Store UserStore
}

func NewUserResource(store UserStore) *UserResource {
	return &UserResource{
		Store: store,
	}
}
func (rs *UserResource) router(r *echo.Group) {
	log.Println("Inside Organization Router")
	r.GET("", rs.list, authMw.CheckAuthorization([]string{"owner", "admin"}))
	r.GET("/roles", rs.listRoles, authMw.CheckAuthorization([]string{"owner", "admin"}))
	r.PATCH("/:id", rs.patchUser, authMw.CheckAuthorization([]string{"owner", "admin"}))
}

type listUsersResp struct {
	Users []userResp `json:"users"`
}

type userResp struct {
	authapi.User
	Role authapi.Role `json:"role"`
}

type roleResp struct {
	Roles []authapi.Role `json:"roles"`
}

type patchRequest struct {
	RoleName *string `json:"role,omitempty"`
}

func (rs *UserResource) patchUser(c echo.Context) error {
	log.Println("Inside patchUser()")

	r := new(patchRequest)

	if err := c.Bind(r); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return err
		}
	}

	//Get the url parameter and parse it into UUID
	orgUserUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println(err)
		return c.JSON(ErrUserNotFound.Error.CodeInt, ErrUserNotFound)
	}

	orgUserToBeUpdated := authapi.OrganizationUser{}
	orgUserToBeUpdated.UUID = orgUserUUID

	if r.RoleName != nil { //checking if the role is being changed

		//TODO Cache this
		roles, err := rs.Store.ListRoles() // list all the roles in the DB
		if err != nil {
			log.Println(err)
			if errCode := authapi.ErrorCode(err); errCode != "" {
				return c.JSON(http.StatusInternalServerError, ErrAuth0Unknown)
			}
			return c.JSON(http.StatusInternalServerError, ErrAuth0Unknown)
		}
		roleFound := false // checking the role is a valid type
		for _, role := range roles {
			if strings.ToUpper(role.Name) == strings.ToUpper(*r.RoleName) {
				orgUserToBeUpdated.RoleID = int(role.AccessLevel)
				roleFound = true
				break
			}
		}
		if !roleFound {
			return c.JSON(http.StatusUnprocessableEntity, ErrRoleNotFound)
		}
	}
	//role found if it made it here
	err = rs.Store.Update(orgUserToBeUpdated)
	if err != nil {
		log.Println(err)
		return c.JSON(ErrInternal.Error.CodeInt, ErrInternal)
	}
	return c.NoContent(http.StatusOK)
}

func (rs *UserResource) listRoles(c echo.Context) error {
	log.Println("Inside list()")

	roles, err := rs.Store.ListRoles()

	if err != nil {
		log.Println(err)
		if errCode := authapi.ErrorCode(err); errCode != "" {
			return c.JSON(http.StatusInternalServerError, ErrAuth0Unknown)
		}
		return c.JSON(http.StatusInternalServerError, ErrAuth0Unknown)
	}

	rResp := roleResp{}
	rResp.Roles = roles

	return c.JSON(http.StatusOK, rResp)

}

func (rs *UserResource) list(c echo.Context) error {
	log.Println("Inside list(first)")
	orgID, _ := strconv.Atoi(c.Get("orgID").(string))
	o := authapi.Organization{}
	o.ID = orgID

	orgUsers, err := rs.Store.List(&o)

	if err != nil {
		log.Println(err)
		if errCode := authapi.ErrorCode(err); errCode != "" {
			return c.JSON(http.StatusInternalServerError, ErrAuth0Unknown)
		}
		return c.JSON(http.StatusInternalServerError, ErrAuth0Unknown)
	}

	/*
		loop through all the OrgUsers and convert them into a "flattened" version.
		I'm not seeing a circumstance when the response should be an object like this:
		OrgUser { User {	} }
	*/

	var usersSlice []userResp
	for _, tempOrgUser := range orgUsers { //
		if tempOrgUser.Active {
			var tempUser = userResp{}
			tempUser.UUID = tempOrgUser.UUID
			tempUser.FirstName = tempOrgUser.User.FirstName
			tempUser.LastName = tempOrgUser.User.LastName
			tempUser.Address = tempOrgUser.User.Address
			tempUser.Email = tempOrgUser.User.Email
			tempUser.Mobile = tempOrgUser.User.Mobile
			tempUser.Phone = tempOrgUser.User.Phone
			tempUser.Role = *tempOrgUser.Role
			tempUser.Active = tempOrgUser.Active
			usersSlice = append(usersSlice, tempUser)
		}
	}
	resp := listUsersResp{
		Users: usersSlice,
	}
	return c.JSON(http.StatusOK, resp)

}
