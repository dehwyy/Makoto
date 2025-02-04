package twirp

import (
	"context"
	"fmt"

	"github.com/dehwyy/makoto/apps/gateway/services"
	"github.com/dehwyy/makoto/libs/grpc/generated/auth"
	"github.com/dehwyy/makoto/libs/grpc/generated/general"
	"github.com/dehwyy/makoto/libs/grpc/generated/user"
	tw "github.com/twitchtv/twirp"
)

type TwirpAuthorizationService struct {
	ReadHeader             func(context.Context) (string, error)
	SetAuthorizationHeader func(context.Context, string) error

	client            auth.AuthRPC
	userServiceClient user.UserRPC
}

func NewAuthorizationService(auth_service_url, user_service_url string, args TwirpAuthorizationService) auth.TwirpServer {
	return auth.NewAuthRPCServer(&TwirpAuthorizationService{
		ReadHeader: args.ReadHeader,

		client: services.NewAuthorizationService(services.AuthorizationService{
			AuthorizationServiceUrl: auth_service_url,
		}),
		userServiceClient: services.NewUserService(services.UserService{
			UserServiceUrl: user_service_url,
		}),
	}, tw.WithServerPathPrefix("/authorization"))
}

func (s *TwirpAuthorizationService) SignUp(ctx context.Context, req *auth.SignUpRequest) (*auth.AuthResponse, error) {
	response, err := s.client.SignUp(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.userServiceClient.CreateUser(ctx, &user.CreateUserPayload{
		UserId: response.UserId,
	})
	if err != nil {
		return nil, err
	}

	if err = s.set_token(ctx, response.Token); err != nil {
		return nil, err
	}

	new_response := &auth.AuthResponse{
		UserId:   response.UserId,
		Username: response.Username,
	}
	return new_response, nil
}

func (s *TwirpAuthorizationService) SignIn(ctx context.Context, req *auth.SignInRequest) (*auth.AuthResponse, error) {
	// if not by credentials && not by oauth2 && token is empty-> attach token
	if req.GetCredentials() == nil && req.GetOauth2() == nil && req.GetToken() == "" {
		token, _ := s.ReadHeader(ctx)
		req = &auth.SignInRequest{
			AuthMethod: &auth.SignInRequest_Token{
				Token: token,
			},
		}
	}

	response, err := s.client.SignIn(ctx, req)
	if err != nil {
		return nil, err
	}

	// if account was created -> create account in UserInfo (when oauth2)
	if response.IsCreated {
		_, err = s.userServiceClient.CreateUser(ctx, &user.CreateUserPayload{
			UserId: response.UserId,
		})
		if err != nil {
			return nil, err
		}
	}

	if err = s.set_token(ctx, response.Token); err != nil {
		return nil, err
	}

	new_response := &auth.AuthResponse{
		UserId:   response.UserId,
		Username: response.Username,
	}

	return new_response, nil
}

func (s *TwirpAuthorizationService) IsUniqueUsername(ctx context.Context, req *auth.IsUniqueUsernamePayload) (*auth.IsUnique, error) {
	response, err := s.client.IsUniqueUsername(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TwirpAuthorizationService) VerifyUserEmail(ctx context.Context, req *general.UserId) (*general.IsSuccess, error) {
	response, err := s.client.VerifyUserEmail(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TwirpAuthorizationService) ChangePassword(ctx context.Context, req *auth.ChangePasswordPayload) (*general.IsSuccess, error) {
	response, err := s.client.ChangePassword(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TwirpAuthorizationService) Logout(ctx context.Context, req *general.UserId) (*general.IsSuccess, error) {
	response, err := s.client.Logout(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TwirpAuthorizationService) set_token(ctx context.Context, token string) error {
	return tw.SetHTTPResponseHeader(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
}
