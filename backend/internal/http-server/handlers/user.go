package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	daecv1 "github.com/Prrromanssss/DAEC-fullstack/internal/protos/gen/go/daec"
	"github.com/Prrromanssss/DAEC-fullstack/internal/storage"
)

// HandlerLoginUser is a http.Handler to login user.
func HandlerLoginUser(
	log *slog.Logger,
	dbCfg *storage.Storage,
	grpcClient daecv1.AuthClient,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerLoginUser"

		log := log.With(
			slog.String("fn", fn),
		)

		type parametrs struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parametrs{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing JSON: %v", err))
			return
		}

		loginResponse, err := grpcClient.Login(r.Context(), &daecv1.LoginRequest{
			Email:    params.Email,
			Password: params.Password,
		})
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("can't login user: %v", err))
			return
		}

		respondWithJson(log, w, 200, loginResponse)
	}
}

// HandlerRegisterNewUser is a http.Handler to register new user.
func HandlerRegisterNewUser(
	log *slog.Logger,
	dbCfg *storage.Storage,
	grpcClient daecv1.AuthClient,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.HandlerLoginUser"

		log := log.With(
			slog.String("fn", fn),
		)

		type parametrs struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parametrs{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("error parsing JSON: %v", err))
			return
		}

		registerResponse, err := grpcClient.Register(r.Context(), &daecv1.RegisterRequest{
			Email:    params.Email,
			Password: params.Password,
		})
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("can't register new user: %v", err))
			return
		}

		err = dbCfg.Queries.NewOperationsForUser(r.Context(), int32(registerResponse.UserId))
		if err != nil {
			respondWithError(log, w, 400, fmt.Sprintf("can't create new operations for user: %v", err))
			return
		}

		respondWithJson(log, w, 200, registerResponse)
	}
}
