package main

import (
	"ad_integration/internal/service"
	ldap "ad_integration/internal/service/ldap"
	"context"
	"fmt"
	"os"

	"github.com/k0kubun/pp"
)

var login string
var password string

type MockClient struct {
}

func (m MockClient) Search(
	ctx context.Context,
	login string,
	filter string,
	attributes []string,
) (*ldap.RawUser, error) {
	mockRawUser := &ldap.RawUser{
		DN: "dsadasda",
		Attributes: map[string][]string{
			"cn":             {"a.khafizov"},
			"sAMAccountName": {"a.khafizov"},
			"memberOf": {"CN=Admins,DC=tp,DC=local",
				"CN=Users,DC=tp,DC=local"}},
	}

	return mockRawUser, nil
}

func (m MockClient) BindUser(
	login string,
	password string,
) error {
	return nil
}

func main() {
	url := os.Getenv("URL")
	baseDN := os.Getenv("BASEDN")
	attributes := []string{
		"cn",
		"memberOf",       // Тут будет лежать твоя lab-test-admins
		"sAMAccountName", // Тут будет лежать lab-admin
	}
	useTLS := true
	serverName := "DC01.tp.local"

	cfg := ldap.LoadConfig(url, baseDN, useTLS, attributes, serverName)

	/*client, err := ldap.NewLDAPConnection(cfg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}*/

	mockClient := MockClient{}

	s := service.NewAuthService(mockClient)

	login = os.Getenv("LOGIN")
	password = os.Getenv("PASSW")
	fmt.Println(login, password)

	rawUser, err := s.Authenticate(context.Background(), login, password, cfg.Attributes)
	if err != nil {
		fmt.Println("Fail to Fetch User's Details:", err)
		return
	}
	pp.Println(rawUser)

	/*ctx := context.Background()
	addressDB := os.Getenv("CONN_STRING")

	conn, err := repository.Connection(ctx, addressDB)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	fmt.Println("connection is succeded")

	userID, err := repository.SyncUser(ctx, conn, userDetails)
	if err != nil {
		fmt.Println("Ошибка при сихронизации user:", err)
		return
	}

	if err := repository.SyncGroups(ctx, conn, userDetails.Groups); err != nil {
		fmt.Println("Ошибка при синхронизации групп:", err)
		return
	}

	if err := repository.RefreshUserRoles(ctx, conn, userID, userDetails.Groups); err != nil {
		fmt.Println("Ошибка при обновление ролей пользователя:", err)
		return
	}*/
}

/*testGroupsFromAD := []string{
	"CN=lab-test-admins,OU=Groups,DC=tp,DC=local",
	"CN=all-staff,OU=Global,DC=tp,DC=local",
	"CN=vpn-users,OU=Access,DC=tp,DC=local",
}
*/
