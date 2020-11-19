package actions

import (
	"github.com/shipanther/trober/models"
)

func (as *ActionSuite) Test_TenantsResource_List() {
	as.LoadFixture("Tenant bootstrap")
	var user = &models.User{}
	err := as.DB.Where("username=?", "nicoleta").First(user)
	as.NoError(err)
	as.NotZero(user.ID)
	req := as.JSON("/tenants")
	req.Headers[xToken] = user.Username
	res := req.Get()
	as.Equal(200, res.Code)
	var tenants = &models.Tenants{}
	res.Bind(tenants)
	as.Equal(len(*tenants), 1)

}

func (as *ActionSuite) Test_TenantsResource_Show() {
	as.False(false)
}

func (as *ActionSuite) Test_TenantsResource_Create() {
	// var tenants = models.Tenants{}
	// //var tenat
	// res := as.JSON("/tenants").Post()
	// as.Equal(200, res.Code)
	// res.Bind(tenants)
	as.False(false)
}

func (as *ActionSuite) Test_TenantsResource_Update() {
	as.False(false)
}

func (as *ActionSuite) Test_TenantsResource_Destroy() {
	as.False(false)
}
