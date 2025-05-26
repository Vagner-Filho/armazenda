package field_model

import (
	entity_public "armazenda/entity/public"
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4" // Using pgxmock as an example
	"github.com/shopspring/decimal"    // Assuming this is the decimal library used
)

// Mock pgx.Conn (or use a library like pgxmock)
// For simplicity, this example will use pgxmock.

func TestFieldModel_AddField(t *testing.T) {
	// Mock successful database connection
	mockConn, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockConn.Close(context.Background())

	fm := &fieldModel{conn: mockConn.Conn()}

	// --- Test Case 1: Successful Field Addition ---
	t.Run("Success", func(t *testing.T) {
		inputField := entity_public.Field{
			Name:     "Test Field 1",
			Farm:     101,
			Hectares: decimal.NewFromFloat(10.5),
		}
		expectedID := uint16(1)
		expectedName := "Test Field 1" // Name returned from DB

		// Expect the QueryRow call
		rows := pgxmock.NewRows([]string{"id", "name"}).AddRow(expectedID, expectedName)
		mockConn.ExpectQuery("INSERT INTO field").
			WithArgs(inputField.Name, inputField.Farm, inputField.Hectares).
			WillReturnRows(rows)

		resultField, modelErr := fm.AddField(inputField)

		if modelErr != nil {
			t.Errorf("Expected no error, but got: %v", modelErr)
		}
		if resultField.Id != expectedID {
			t.Errorf("Expected ID %d, but got %d", expectedID, resultField.Id)
		}
		if resultField.Name != expectedName {
			t.Errorf("Expected Name '%s', but got '%s'", expectedName, resultField.Name)
		}
		// Note: The original function doesn't populate Farm and Hectares from the RETURNING clause.
		// It uses the uninitialized local variables `farm` and `ha`.
		// If this is intended, the test should reflect that (e.g., expect zero values).
		// If it's a bug, the test will highlight it. For this test, we'll assume current behavior.
		if resultField.Farm != 0 { // Assuming farm is not returned and initialized to 0
			t.Errorf("Expected Farm 0, but got %d", resultField.Farm)
		}
		if !resultField.Hectares.IsZero() { // Assuming Hectares is not returned and initialized to decimal.Zero
			t.Errorf("Expected Hectares 0, but got %s", resultField.Hectares.String())
		}

		// Ensure all expectations were met
		if err := mockConn.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	// --- Test Case 2: Unique Constraint Violation ---
	t.Run("ErrorUniqueViolation", func(t *testing.T) {
		inputField := entity_public.Field{
			Name:     "Duplicate Field",
			Farm:     102,
			Hectares: decimal.NewFromFloat(20.0),
		}
		pgErr := &pgconn.PgError{
			Code:    pgerrcode.UniqueViolation,
			Message: "unique constraint violation", // Example message
		}

		mockConn.ExpectQuery("INSERT INTO field").
			WithArgs(inputField.Name, inputField.Farm, inputField.Hectares).
			WillReturnError(pgErr)

		resultField, modelErr := fm.AddField(inputField)

		if modelErr == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if modelErr.Message != "Já existe um talhão com este nome" {
			t.Errorf("Expected error message 'Já existe um talhão com este nome', but got '%s'", modelErr.Message)
		}
		if modelErr.IsServerErr {
			t.Error("Expected IsServerErr to be false for unique violation")
		}
		if (resultField != entity_public.Field{}) {
			t.Errorf("Expected empty entity_public.Field on error, but got %+v", resultField)
		}

		if err := mockConn.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	// --- Test Case 3: Other Database Error on Query/Scan ---
	t.Run("ErrorOtherDatabaseError", func(t *testing.T) {
		inputField := entity_public.Field{
			Name:     "FieldWithError",
			Farm:     103,
			Hectares: decimal.NewFromFloat(5.0),
		}
		dbErr := errors.New("some other database error") // Generic error

		mockConn.ExpectQuery("INSERT INTO field").
			WithArgs(inputField.Name, inputField.Farm, inputField.Hectares).
			WillReturnError(dbErr)

		// Reset mock logger if it's stateful or capture its output
		// For simplicity, we assume model_error.Logger doesn't panic here.
		// You might want to mock or spy on model_error.Logger if its side effects are critical to test.

		resultField, modelErr := fm.AddField(inputField)

		if modelErr == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if modelErr.Message != "Falhamos ao adicionar o talhão" {
			t.Errorf("Expected error message 'Falhamos ao adicionar o talhão', but got '%s'", modelErr.Message)
		}
		if !modelErr.IsServerErr {
			t.Error("Expected IsServerErr to be true for other database errors")
		}
		if (resultField != entity_public.Field{}) {
			t.Errorf("Expected empty entity_public.Field on error, but got %+v", resultField)
		}

		if err := mockConn.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	// --- Test Case 4: Scan Error (not pgconn.PgError) ---
	// This can happen if, for example, the number of columns returned by RETURNING
	// doesn't match the number of scan destinations, or a type mismatch that Scan can't handle.
	t.Run("ErrorScanNonPgError", func(t *testing.T) {
		inputField := entity_public.Field{
			Name:     "FieldScanError",
			Farm:     104,
			Hectares: decimal.NewFromFloat(12.0),
		}
		// Simulate a scan error by returning rows that would cause Scan to fail,
		// but the error itself is not a *pgconn.PgError.
		// For example, pgx.ErrNoRows is a common one if QueryRow doesn't find a row (though RETURNING usually guarantees one on success).
		// More directly, we can just make the WillReturnError return a generic error.
		scanRelatedErr := errors.New("simulated scan error, not a PgError")

		mockConn.ExpectQuery("INSERT INTO field").
			WithArgs(inputField.Name, inputField.Farm, inputField.Hectares).
			WillReturnError(scanRelatedErr) // This error will be wrapped by QueryRow().Scan()

		resultField, modelErr := fm.AddField(inputField)

		if modelErr == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if modelErr.Message != "Falhamos ao adicionar o talhão" {
			t.Errorf("Expected error message 'Falhamos ao adicionar o talhão', but got '%s'", modelErr.Message)
		}
		if !modelErr.IsServerErr {
			t.Error("Expected IsServerErr to be true for this type of scan error")
		}
		if (resultField != entity_public.Field{}) {
			t.Errorf("Expected empty entity_public.Field on error, but got %+v", resultField)
		}

		if err := mockConn.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	// --- Potential Bug Identification / Test for current behavior ---
	// The original function declares local variables farm and ha but returns them
	// without populating them from the database response in the success case.
	// The RETURNING clause only gets id and name.
	t.Run("VerifyReturnValuesFarmHectaresNotPopulatedOnSuccess", func(t *testing.T) {
		inputField := entity_public.Field{
			Name:     "Test Field Unpopulated",
			Farm:     999,                        // Input farm value
			Hectares: decimal.NewFromFloat(99.9), // Input hectares value
		}
		expectedID := uint16(5)
		expectedName := "Test Field Unpopulated"

		rows := pgxmock.NewRows([]string{"id", "name"}).AddRow(expectedID, expectedName)
		mockConn.ExpectQuery("INSERT INTO field").
			WithArgs(inputField.Name, inputField.Farm, inputField.Hectares).
			WillReturnRows(rows)

		resultField, modelErr := fm.AddField(inputField)

		if modelErr != nil {
			t.Errorf("Expected no error, but got: %v", modelErr)
		}
		// Verify that the returned Farm and Hectares are the zero values of their types,
		// not the input values, because they are not scanned from the DB response.
		if resultField.Farm != 0 { // uint32 zero value
			t.Errorf("Expected returned Farm to be 0 (uninitialized from DB), but got %d", resultField.Farm)
		}
		if !resultField.Hectares.IsZero() { // decimal.Decimal zero value
			t.Errorf("Expected returned Hectares to be 0 (uninitialized from DB), but got %s", resultField.Hectares.String())
		}
		if resultField.Id != expectedID {
			t.Errorf("Expected ID %d, but got %d", expectedID, resultField.Id)
		}
		if resultField.Name != expectedName {
			t.Errorf("Expected Name '%s', but got '%s'", expectedName, resultField.Name)
		}

		if err := mockConn.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

// Helper to setup fieldModel if needed in multiple test files or more complex scenarios
// func newTestFieldModel(db *pgx.Conn) *fieldModel {
// return &fieldModel{conn: db}
// }

// You might need to mock model_error.Logger if it has side effects
// you want to control or verify (e.g., writing to a specific output).
// For instance, you could replace it with a mock during tests:
//
// var originalLoggerFunc = model_error.Logger
//
// func setupMockLogger() (func(), *[]string) {
//   var logs []string
//   model_error.Logger = func(conn *pgx.Conn, msg string) {
//     logs = append(logs, msg)
//   }
//   cleanup := func() {
//     model_error.Logger = originalLoggerFunc
//   }
//   return cleanup, &logs
// }
//
// And then in your test:
// cleanup, logs := setupMockLogger()
// defer cleanup()
// ... your test logic ...
// if len(*logs) == 0 { t.Error("Expected logger to be called") }
