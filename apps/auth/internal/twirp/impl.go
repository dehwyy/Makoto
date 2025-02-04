package twirp

import (
	"context"
	"errors"

	"github.com/dehwyy/makoto/apps/auth/internal/oauth2"
	"github.com/dehwyy/makoto/apps/auth/internal/repository"
	"github.com/dehwyy/makoto/libs/config"
	"github.com/dehwyy/makoto/libs/database/models"
	"github.com/dehwyy/makoto/libs/grpc/generated/auth"
	"github.com/dehwyy/makoto/libs/grpc/generated/general"
	"github.com/dehwyy/makoto/libs/logger"
	"github.com/google/uuid"
	tw "github.com/twitchtv/twirp"
	oauth2lib "golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Server struct {
	// repository
	token_repository *repository.TokenRepository
	user_repository  *repository.UserRepository

	// oauth
	oauth2 *oauth2.OAuth2Creator

	//
	l logger.Logger
}

func NewTwirpServer(db *gorm.DB, config *config.Config, l logger.Logger) auth.TwirpServer {
	token_repo := repository.NewTokenRepository(db, l, config.JwtSecret)
	user_repo := repository.NewUserRepository(db, l)

	return auth.NewAuthRPCServer(&Server{
		// repo
		token_repository: token_repo,
		user_repository:  user_repo,

		// oauth
		oauth2: oauth2.NewOAuth2Creator(token_repo, config, l),

		// logger
		l: l,
	})
}

// SignUp creates a new user account and returns an authentication response. (credentials only)
func (s *Server) SignUp(ctx context.Context, req *auth.SignUpRequest) (*auth.AuthResponse, error) {
	if req.Username == "" {
		return nil, tw.InvalidArgumentError("username", "username cannot be empty")
	} else if req.Email == "" {
		return nil, tw.InvalidArgumentError("email", "email cannot be empty")
	} else if req.Password == "" {
		return nil, tw.InvalidArgumentError("password", "password cannot be empty")
	}

	user_uuid := uuid.New()

	if err := s.user_repository.CreateUser(repository.CreateUserPayload{
		ID:       user_uuid,
		Id:       user_uuid.String(),
		Email:    req.Email,
		Username: req.Username,
		Picture:  "",
		Password: req.Password,
		Provider: repository.ProviderLocal,
	}); err != nil {

		s.l.Errorf("create user: %v", err)

		if errors.Is(err, repository.UserAlreadyExists) {
			return nil, tw.InvalidArgumentError("username", err.Error())
		} else {
			return nil, tw.InternalError(err.Error())
		}
	}

	token, err := s.token_repository.CreateToken(user_uuid, req.Username)
	if err != nil {
		s.l.Errorf("create token: %v", err)
		return nil, tw.InternalError(err.Error())
	}

	return &auth.AuthResponse{
		Token:     token,
		Username:  req.Username,
		UserId:    user_uuid.String(),
		IsCreated: true,
	}, nil
}

// SignIn authenticates a user and returns an authentication response.
//
// Parameters:
// - ctx: the context.Context object for handling deadlines, cancellation signals, and other request-scoped values.
// - req: a pointer to the auth.SignInRequest object containing the user's sign-in request.
//
// Returns:
// - An auth.AuthResponse pointer that contains the authentication response.
// - An error object if any error occurs during the authentication process.
func (s *Server) SignIn(ctx context.Context, req *auth.SignInRequest) (*auth.AuthResponse, error) {

	token := req.GetToken()
	var found_token *models.UserToken
	var found_user *models.UserData

	// token authorization
	if token != "" && req.GetOauth2() == nil && req.GetCredentials() == nil {

		// try to find this token in db
		token_db, err := s.token_repository.GetToken(token)
		if err != nil {
			s.l.Errorf("get token: %v", err)
			return nil, tw.PermissionDenied.Error(err.Error())
		}

		user, err := s.user_repository.GetUserById(repository.GetUserPayload{
			Id: &token_db.UserId,
		})
		if err != nil {
			s.l.Errorf("get user: %v", err)
			return nil, tw.PermissionDenied.Error(err.Error())
		}

		if user.Provider == repository.ProviderLocal {

			token = token_db.AccessToken

			// if validator returns error -> regenerate token
			if s.token_repository.ValidateToken(token_db.AccessToken) != nil {
				token, err = s.token_repository.UpdateToken(token_db.UserId)
				if err != nil {
					s.l.Errorf("update token: %v", err)
					return nil, tw.InternalError(err.Error())
				}
			}

			return &auth.AuthResponse{
				Token:     token,
				Username:  user.Username,
				UserId:    user.ID.String(),
				IsCreated: false,
			}, nil
		}

		found_user = user
		found_token = token_db
	}

	// OAuth2 SignIn flow
	if oauth2_input := req.GetOauth2(); oauth2_input != nil || found_user != nil {

		// is it ok? xd
		var oauth2_inst *oauth2.OAuth2
		var token_db *oauth2lib.Token
		var status oauth2.TokenStatus

		// direct request
		if oauth2_input != nil {
			oauth2_inst = s.oauth2.NewOAuth2(oauth2_input.Provider)
			token_db, status = oauth2_inst.GetToken(token, oauth2_input.GetCode())

		} else { // only token request (it proceeds here from above (scope `if request is Empty`))
			oauth2_inst = s.oauth2.NewOAuth2(string(found_user.Provider))
			token_db, status = oauth2_inst.GetToken(found_token.AccessToken, "")
		}

		switch status {
		case oauth2.Redirect:
			return nil, tw.NewError(tw.Unauthenticated, "use provider credentials")
		case oauth2.InternalError:
			return nil, tw.NewError(tw.Internal, "internal error")
		}

		response, err := oauth2_inst.GetUserByToken(token_db)
		if err != nil {
			return nil, tw.InternalErrorf("internal error %v", err.Error())
		}

		found_user, err := s.user_repository.GetUserByProviderId(response.Id)

		//  if no user was found => create new user + new token in db => return
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// generate new uuid for user
			user_uuid := uuid.New()

			// creating payload from response and other data
			createUserPayload := repository.CreateUserPayload{
				ID:       user_uuid,
				Id:       response.Id,
				Email:    response.Email,
				Username: response.Username,
				Picture:  response.Picture, // it would remove fixed size of image in this case `s96-c`
				Provider: models.AuthProvider(oauth2_inst.GetProviderName()),
				Password: "", // no password actually
			}

			err = s.user_repository.CreateUser(createUserPayload)
			if err != nil {
				s.l.Errorf("create user: %v", err)
				return nil, err
			}

			err = s.token_repository.CreateTokenByOAuth2Token(user_uuid, token_db)
			if err != nil {
				s.l.Errorf("create token: %v", err)
				return nil, err
			}

			return &auth.AuthResponse{
				Token:     token_db.AccessToken,
				Username:  response.Username,
				UserId:    user_uuid.String(),
				IsCreated: true,
			}, err
		}

		// else if user was found => update token in db
		err = s.token_repository.UpdateTokenByOAuth2Token(found_user.ID, token_db)
		if err != nil {
			s.l.Errorf("save token: %v", err)
			return nil, err
		}

		return &auth.AuthResponse{
			Token:     token_db.AccessToken,
			Username:  found_user.Username,
			UserId:    found_user.ID.String(),
			IsCreated: false,
		}, nil
	}

	// ! By credentials
	credentials := req.GetCredentials()

	userId, username, err := s.user_repository.ValidateUser(repository.ValidateUserPayload{
		// ? either Username or Email would/should be provided
		Username: credentials.GetUsername(),
		Email:    credentials.GetEmail(),
		// always
		Password: credentials.GetPassword(),
	})

	if errors.Is(err, repository.UserNotFound) {
		return nil, tw.NewError(tw.Unauthenticated, "user with provided credentials wasn't found")

	} else if errors.Is(err, repository.UserWrongPassword) {
		return nil, tw.NewError(tw.Unauthenticated, "wrong password")

	} // would not return unnamed error => no simple check for nil

	token, err = s.token_repository.UpdateToken(*userId)
	if err != nil {
		return nil, tw.InternalError(err.Error())
	}

	return &auth.AuthResponse{
		Token:     token,
		Username:  username,
		UserId:    userId.String(),
		IsCreated: false,
	}, nil
}

// IsUniqueUsername checks if a username is unique.
//
// Takes a context.Context and an *auth.IsUniqueUsernamePayload as parameters.
// Returns an *auth.IsUnique and an error.
func (s *Server) IsUniqueUsername(ctx context.Context, req *auth.IsUniqueUsernamePayload) (*auth.IsUnique, error) {
	_, err := s.user_repository.GetUserByUsername(req.Username)
	// if user wasn't found -> it will return error (instanceof gorm.ErrRecordNotFound)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &auth.IsUnique{IsUnique: true}, nil
	}

	// internal error
	if err != nil {
		return nil, tw.InternalError(err.Error())
	}

	return &auth.IsUnique{IsUnique: false}, nil
}

// TODO: untested yet
// VerifyUserEmail verifies the user email.
//
// It takes a context.Context and a *general.UserId as parameters.
// It returns a *general.IsSuccess and an error.
func (s *Server) VerifyUserEmail(ctx context.Context, req *general.UserId) (*general.IsSuccess, error) {
	user_id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, tw.InternalError(err.Error())
	}

	err = s.user_repository.VerifyUserEmail(user_id)
	if err != nil {
		return nil, tw.InternalError(err.Error())
	}

	return &general.IsSuccess{IsSuccess: true}, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *auth.ChangePasswordPayload) (*general.IsSuccess, error) {
	user_id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, tw.InvalidArgumentError("invalid userId", err.Error())
	}

	_, _, err = s.user_repository.ValidateUser(repository.ValidateUserPayload{
		UserId:   user_id,
		Password: req.OldPassword,
	})
	if err != nil {
		if errors.Is(err, repository.UserNotFound) {
			return nil, tw.NewError(tw.NotFound, "user with provided credentials wasn't found")
		} else if errors.Is(err, repository.UserWrongPassword) {
			return nil, tw.NewError(tw.Unauthenticated, "wrong password")
		}
	}
	// -> validation passed -> allow change password

	err = s.user_repository.UpdateUserPassword(user_id, req.NewPassword)
	if err != nil {
		return nil, tw.InternalError(err.Error())
	}

	return &general.IsSuccess{IsSuccess: true}, nil
}

func (s *Server) Logout(ctx context.Context, req *general.UserId) (*general.IsSuccess, error) {
	user_id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, tw.InvalidArgumentError("invalid userId", err.Error())
	}

	err = s.token_repository.DeleteTokenByUserId(user_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, tw.NotFoundError("token with provided userId wasn't found")
		}
		return nil, tw.InternalError(err.Error())
	}

	return &general.IsSuccess{IsSuccess: true}, nil
}
