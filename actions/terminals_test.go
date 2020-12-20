package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_TerminalsList() {
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
	newTerminal := as.createTerminal("terminal", models.TerminalTypePort, firmino.TenantID, firmino.ID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/terminals")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
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

func (as *ActionSuite) Test_TerminalsListFilter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			terminalType := models.TerminalTypePort
			if i%2 == 0 {
				terminalType = models.TerminalTypeRail
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
		as.Equal(string(models.TerminalTypePort), v.Type)
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

func (as *ActionSuite) Test_TerminalsShow() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusOK},
		{"firmino", http.StatusOK},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusOK},
		{"klopp", http.StatusOK},
		{"adidas", http.StatusOK},
	}
	lewin := as.getLoggedInUser("lewin")
	firmino := as.getLoggedInUser("firmino")
	as.NotEqual(firmino.TenantID, lewin.TenantID)

	var terminals = []*models.Terminal{as.createTerminal("term1", models.TerminalTypePort, firmino.TenantID, firmino.ID),
		as.createTerminal("term2", models.TerminalTypeRail, lewin.TenantID, lewin.ID)}
	as.NotEqual(terminals[0].TenantID, terminals[1].TenantID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			for _, v := range terminals {
				req := as.setupRequest(user, fmt.Sprintf("/terminals/%s", v.ID))
				res := req.Get()
				if v.TenantID == user.TenantID || user.IsSuperAdmin() {
					as.Equal(test.responseCode, res.Code)
				} else {
					as.Equal(http.StatusNotFound, res.Code)
				}
				if res.Code == http.StatusOK {
					var terminal = models.Terminal{}
					res.Bind(&terminal)
					as.Equal(v.Name, terminal.Name)
					as.Equal(v.Type, terminal.Type)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_TerminalsCreate() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusCreated},
		{"firmino", http.StatusCreated},
		{"rodriguez", http.StatusCreated},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"klopp", http.StatusCreated},
		{"adidas", http.StatusNotFound},
	}
	var firmino = as.getLoggedInUser("firmino")
	for i, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			terminalType := models.TerminalTypePort
			if i%2 == 0 {
				terminalType = models.TerminalTypeRail
			}
			newTerminal := models.Terminal{Name: user.Username, Type: string(terminalType), TenantID: firmino.TenantID}
			req := as.setupRequest(user, "/terminals")
			res := req.Post(newTerminal)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var terminal = models.Terminal{}
				res.Bind(&terminal)
				as.Equal(newTerminal.Name, terminal.Name)
				as.Equal(newTerminal.Type, terminal.Type)
				if user.IsSuperAdmin() {
					as.Equal(firmino.TenantID, terminal.TenantID)
				} else {
					as.Equal(user.TenantID, terminal.TenantID)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_TerminalsUpdate() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusOK},
		{"firmino", http.StatusOK},
		{"rodriguez", http.StatusNotFound},
		{"coutinho", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"klopp", http.StatusOK},
		{"adidas", http.StatusNotFound},
		{"nike", http.StatusNotFound},
	}
	var firmino = as.getLoggedInUser("firmino")
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			terminalType := models.TerminalTypePort
			newTerminal := as.createTerminal(user.Username, terminalType, firmino.TenantID, firmino.ID)
			req := as.setupRequest(user, fmt.Sprintf("/terminals/%s", newTerminal.ID))
			// Try to update ID and tenant ID. Expect these calls to be excluded at update
			updatedTerminal := models.Terminal{Name: fmt.Sprintf("not%s", test.username), Type: string(models.TerminalTypeRail), ID: user.ID, TenantID: user.ID}
			res := req.Put(updatedTerminal)
			as.Equal(test.responseCode, res.Code)
			var dbTerminal = *newTerminal
			err := as.DB.Reload(&dbTerminal)
			as.Nil(err)
			if res.Code == http.StatusOK {
				var terminal = models.Terminal{}
				res.Bind(&terminal)
				as.Equal(updatedTerminal.Name, terminal.Name)
				as.Equal(updatedTerminal.Type, terminal.Type)
				as.Equal(newTerminal.ID, terminal.ID)
				as.Equal(dbTerminal.Name, terminal.Name)
			} else {
				// Ensure update did not succeed
				as.Equal(dbTerminal.Name, newTerminal.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TerminalsDestroy() {
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
			newTerminal := as.createTerminal(name, models.TerminalTypePort, firmino.TenantID, firmino.ID)

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/terminals/%s", newTerminal.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
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
