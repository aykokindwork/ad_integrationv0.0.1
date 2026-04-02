package service

import (
	errors2 "ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"ad_integration/internal/auth/repository/repositorymocks"
	servicemocks "ad_integration/internal/auth/service/serviceMocks"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
)

func TestAuthenticate(t *testing.T) {
	type Authenticate struct {
		res *model.LDAPUser

		bindExpect   bool
		searchExpect bool

		bindErr   error
		searchErr error

		wantErr   error
		searchRes *model.RawUser
	}

	tests := []struct {
		name         string
		Authenticate Authenticate
		wantErr      error
	}{
		{
			name: "Success",
			Authenticate: Authenticate{
				res: &model.LDAPUser{
					Username: "some",
					Email:    "some@example.com",
					Groups: []string{
						"Group1",
					},
				},
				searchRes: &model.RawUser{
					DN: "some",
					Attributes: map[string][]string{
						attrSAMAccountName:    {"some"},
						attrUserPrincipalName: {"some@example.com"},
						attrMemberOf:          {"CN=Group1,DC=tp"},
					},
				},

				bindErr:      nil,
				bindExpect:   true,
				searchExpect: true,
				searchErr:    nil,
				wantErr:      nil,
			},
		},
		{
			name: "Fail_BindUser_invalidCred",
			Authenticate: Authenticate{
				res:        nil,
				bindExpect: true,

				searchExpect: false,
				bindErr:      errors2.ErrLdapBind,
				searchErr:    nil,
				searchRes:    nil,
				wantErr:      errors2.ErrLdapBind,
			},
			wantErr: errors2.ErrLdapBind,
		},
		{
			name: "Fail_Search_Invalid",
			Authenticate: Authenticate{
				res:          nil,
				bindExpect:   true,
				searchExpect: true,

				bindErr:   nil,
				searchErr: errors2.ErrLdapSearch,
				wantErr:   errors2.ErrLdapSearch,
			},
			wantErr: errors2.ErrLdapSearch,
		},
		{
			name: "Fail_Len_AttrUserPrincipalName",
			Authenticate: Authenticate{
				res:          nil,
				bindExpect:   true,
				searchExpect: true,
				bindErr:      nil,
				searchErr:    nil,
				wantErr:      errors2.ErrLdapNoEmail,
				searchRes: &model.RawUser{
					DN: "some",
					Attributes: map[string][]string{
						"userPrincipalName": {
							"",
						},
					},
				},
			},

			wantErr: errors2.ErrLdapNoEmail,
		},
		{
			name: "Fail_Len_sAMAccountName",
			Authenticate: Authenticate{
				res:          nil,
				bindExpect:   true,
				searchExpect: true,
				bindErr:      nil,
				searchErr:    nil,
				wantErr:      errors2.ErrLdapNoName,
				searchRes: &model.RawUser{
					DN: "some",
					Attributes: map[string][]string{
						"userPrincipalName": {
							"some@example.com",
						},
						"sAMAccountName": {
							"",
						},
					},
				},
			},

			wantErr: errors2.ErrLdapNoName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			provider := servicemocks.NewMockIdentityProvider(ctrl)
			userRepository := repositorymocks.NewMockUserRepository(ctrl)

			if tt.Authenticate.bindExpect {
				provider.EXPECT().BindUser(
					"some",
					"some",
				).Return(tt.Authenticate.bindErr)
			}

			if tt.Authenticate.searchExpect {
				provider.EXPECT().Search(
					gomock.Any(),
					"(sAMAccountName=some)",
					[]string{"sAMAccountName", "userPrincipalName", "memberOf"},
				).Return(tt.Authenticate.searchRes, tt.Authenticate.searchErr)
			}

			AuthService := NewAuthService(provider, userRepository)

			ldapUser, err := AuthService.Authenticate(context.Background(), "some", "some")

			require.ErrorIs(t, err, tt.Authenticate.wantErr)

			require.Equal(t, tt.Authenticate.res, ldapUser)
		})
	}
}
