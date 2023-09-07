package services

import (
	"context"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"housematee-tgbot/config"
	"log"
)

var (
	gSheets GSheets
)

type IGSheets interface {
	Get(ctx context.Context, spreadsheetId string, readRange string) (*sheets.ValueRange, error)
	Update(ctx context.Context, spreadsheetId string, writeRange string, vr *sheets.ValueRange) (*sheets.UpdateValuesResponse, error)
}

type GSheets struct {
	Svc *sheets.Service
}

func newGSheets(svc *sheets.Service) *GSheets {
	return &GSheets{
		Svc: svc,
	}
}

func InitGSheetsSvc(ctx context.Context, credential config.Credentials) (*GSheets, error) {
	jwtConfig := &jwt.Config{
		Email:      credential.ClientEmail,
		PrivateKey: []byte(credential.PrivateKey),
		Scopes:     []string{sheets.SpreadsheetsScope},
		TokenURL:   credential.TokenURI,
	}

	// Obtain an OAuth2 token for the service account
	tokenSource := jwtConfig.TokenSource(ctx)
	_, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Unable to obtain token: %v", err)
	}

	// Create a new Sheets service with the token
	svc, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}

	gSheets = *newGSheets(svc)
	return &gSheets, nil
}

func (g *GSheets) Get(ctx context.Context, spreadsheetId string, readRange string) (*sheets.ValueRange, error) {
	return g.Svc.Spreadsheets.Values.Get(spreadsheetId, readRange).Context(ctx).Do()
}

func (g *GSheets) Update(ctx context.Context, spreadsheetId string, writeRange string, valueRange *sheets.ValueRange) (*sheets.UpdateValuesResponse, error) {
	return g.Svc.Spreadsheets.Values.Update(spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Context(ctx).Do()
}

func GetGSheetsSvc() GSheets {
	return gSheets
}
