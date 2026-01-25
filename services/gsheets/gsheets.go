package services

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"housematee-tgbot/config"
)

var (
	gSheets GSheets
)

type IGSheets interface {
	Get(ctx context.Context, spreadsheetId string, readRange string) (*sheets.ValueRange, error)
	Update(ctx context.Context, spreadsheetId string, writeRange string, vr *sheets.ValueRange) (*sheets.UpdateValuesResponse, error)
	GetValue(ctx context.Context, spreadsheetId string, readRange string) (string, error)
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
		logrus.Fatalf("unable to obtain token: %v", err)
	}

	// Create a new Sheets service with the token
	svc, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		logrus.Fatalf("unable to create Sheets service: %v", err)
	}

	gSheets = *newGSheets(svc)
	return &gSheets, nil
}

func GetGSheetsSvc() *GSheets {
	return &gSheets
}

func (g *GSheets) Get(ctx context.Context, spreadsheetId string, readRange string) (*sheets.ValueRange, error) {
	return g.Svc.Spreadsheets.Values.Get(spreadsheetId, readRange).Context(ctx).Do()
}

func (g *GSheets) Update(ctx context.Context, spreadsheetId string, writeRange string, valueRange *sheets.ValueRange) (*sheets.UpdateValuesResponse, error) {
	return g.Svc.Spreadsheets.Values.Update(spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Context(ctx).Do()
}

func (g *GSheets) GetValue(ctx context.Context, spreadsheetId string, readRange string) (string, error) {
	resp, err := g.Get(ctx, spreadsheetId, readRange)
	if err != nil {
		return "", err
	}
	if len(resp.Values) == 0 {
		return "", nil
	}
	return resp.Values[0][0].(string), nil
}

// GetSpreadsheet retrieves the spreadsheet metadata including all sheets
func (g *GSheets) GetSpreadsheet(ctx context.Context, spreadsheetId string) (*sheets.Spreadsheet, error) {
	return g.Svc.Spreadsheets.Get(spreadsheetId).Context(ctx).Do()
}

// DuplicateSheet copies a sheet and renames it in a single batch operation
// Returns the new sheet's properties
func (g *GSheets) DuplicateSheet(ctx context.Context, spreadsheetId string, sourceSheetId int64, newTitle string) (*sheets.SheetProperties, error) {
	// Create batch update request with duplicate sheet request
	requests := []*sheets.Request{
		{
			DuplicateSheet: &sheets.DuplicateSheetRequest{
				SourceSheetId:    sourceSheetId,
				NewSheetName:     newTitle,
				InsertSheetIndex: 1, // Insert after the first sheet (Database)
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	resp, err := g.Svc.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	// Get the new sheet properties from the response
	if len(resp.Replies) > 0 && resp.Replies[0].DuplicateSheet != nil {
		return resp.Replies[0].DuplicateSheet.Properties, nil
	}

	return nil, nil
}
