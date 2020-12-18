package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_TerminalsResource_List() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username      string
		responseCode  int
		terminalCount int
	}{
		{"klopp", http.StatusOK, 1},
		{"firmino", http.StatusOK, 1},
		{"mane", http.StatusOK, 1},
		{"salah", http.StatusOK, 1},
		{"nike", http.StatusOK, 1},
		{"coutinho", http.StatusNotFound, 0},
		{"richarlson", http.StatusOK, 0},
		{"rodriguez", http.StatusOK, 0},
		{"lewin", http.StatusOK, 0},
		{"allan", http.StatusNotFound, 0},
		{"adidas", http.StatusOK, 0},
	}
	firmino := as.getLoggedInUser("firmino")
	newTerminal := as.createTerminal("terminal", "Port", firmino.TenantID, firmino.ID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/terminals")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var terminals = models.Terminals{}
				res.Bind(&terminals)
				as.Equal(test.terminalCount, len(terminals))
				if test.terminalCount > 0 {
					as.Equal(newTerminal.Name, terminals[0].Name)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_TerminalsResource_List_Filter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			terminalType := "Port"
			if i%2 == 0 {
				terminalType = "Airport"
			}
			_ = as.createTerminal(fmt.Sprintf("%s-%d", p, i), terminalType, user.TenantID, user.ID)
		}
	}
	req := as.setupRequest(user, "/terminals?name=ਪੰ&type=Port")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var terminals = models.Terminals{}
	res.Bind(&terminals)
	as.Equal(2, len(terminals))
	for _, v := range terminals {
		as.Contains(v.Name, "ਪੰਜਾਬੀ")
		as.Equal("Port", v.Type)
	}
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(klopp.TenantID, user.TenantID)

	as.False(user.IsSuperAdmin())
	req = as.setupRequest(user, fmt.Sprintf("/terminals?tenant_id=%s", klopp.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	terminals = models.Terminals{}
	res.Bind(&terminals)
	as.Equal(0, len(terminals))

	lewin := as.getLoggedInUser("lewin")
	as.NotEqual(klopp.TenantID, lewin.TenantID)
	as.NotEqual(lewin.TenantID, user.TenantID)
	as.True(klopp.IsSuperAdmin())
	req = as.setupRequest(klopp, fmt.Sprintf("/terminals?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	terminals = models.Terminals{}
	res.Bind(&terminals)
	as.Equal(0, len(terminals))

}

func (as *ActionSuite) Test_TerminalsResource_Show() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Create() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Update() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Destroy() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"klopp", http.StatusOK},
		{"firmino", http.StatusOK},
		{"mane", http.StatusOK},
		{"salah", http.StatusNotFound},
		{"nike", http.StatusNotFound},
		{"coutinho", http.StatusNotFound},
		{"richarlson", http.StatusNotFound},
		{"rodriguez", http.StatusNotFound},
		{"lewin", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"adidas", http.StatusNotFound},
	}
	firmino := as.getLoggedInUser("firmino")

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			var name = fmt.Sprintf("terminal%s", test.username)
			newTerminal := as.createTerminal(name, "Port", firmino.TenantID, firmino.ID)

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/terminals/%s", newTerminal.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var terminal = models.Terminal{}
				res.Bind(&terminal)
				as.Equal(name, terminal.Name)
				// Check if actually deleted
				terminal = models.Terminal{}
				err := as.DB.Where("name=?", name).First(&terminal)
				as.Contains(err.Error(), "no rows in result set")
			} else {
				terminal := models.Terminal{}
				err := as.DB.Where("name=?", name).First(&terminal)
				as.Nil(err)
				//Not deleted yet
				as.Equal(name, terminal.Name)
			}
		})
	}
}
