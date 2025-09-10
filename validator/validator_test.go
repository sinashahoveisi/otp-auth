package validator

import (
	"testing"

	"otp-auth/entity"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := New()

	assert.NotNil(t, v)
	assert.NotNil(t, v.validator)
}

func TestValidator_ValidateStruct_Success(t *testing.T) {
	v := New()

	// Test with valid OTP request
	req := entity.SendOTPRequest{
		PhoneNumber: "+1234567890",
	}

	err := v.ValidateStruct(&req)
	assert.NoError(t, err)
}

func TestValidator_ValidateStruct_ValidationError(t *testing.T) {
	v := New()

	// Test with invalid phone number
	req := entity.SendOTPRequest{
		PhoneNumber: "invalid-phone",
	}

	err := v.ValidateStruct(&req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone_number")
}

func TestValidator_ValidateStruct_MissingPhoneNumber(t *testing.T) {
	v := New()

	// Test with missing phone number
	req := entity.SendOTPRequest{}

	err := v.ValidateStruct(&req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone_number")
}

func TestValidator_ValidatePhoneNumber_Valid(t *testing.T) {
	v := New()

	validPhones := []string{
		"+1234567890",
		"+12345678901",
		"+123456789012",
		"+12345678901234",
		"+987654321098765",
		"+19876543210987",
		"+449876543210",
		"+8613912345678",
		"+34612345678",
		"+33612345678",
		"+5511987654321",
		"+61412345678",
		"+4915123456789",
		"+911234567890",
		"+81901234567",
	}

	for _, phone := range validPhones {
		req := entity.SendOTPRequest{PhoneNumber: phone}
		err := v.ValidateStruct(&req)
		assert.NoError(t, err, "Phone number %s should be valid", phone)
	}
}

func TestValidator_ValidatePhoneNumber_Invalid(t *testing.T) {
	v := New()

	invalidPhones := []string{
		"",                      // empty
		"1234567890",            // missing +
		"+0234567890",           // starts with 0 after +
		"+12345",                // too short
		"+123456789012345678",   // too long
		"salamsalam",            // random string
		"+abc1234567890",        // contains letters
		"++1234567890",          // double +
		"+1-234-567-890",        // contains dashes
		"+1 234 567 890",        // contains spaces
		"+1(234)567-890",        // contains parentheses
		"phone_number",          // text
		"123-456-7890",          // US format without +
		"(123) 456-7890",        // US format with parentheses
		"+",                     // just +
		"+1",                    // too short after +
		"+12345678901234567890", // way too long
		"12345678901",           // long but no +
		"+123456789a",           // ends with letter
		"+ 1234567890",          // space after +
	}

	for _, phone := range invalidPhones {
		req := entity.SendOTPRequest{PhoneNumber: phone}
		err := v.ValidateStruct(&req)
		assert.Error(t, err, "Phone number %s should be invalid", phone)
	}
}

func TestValidator_ValidateVerifyOTPRequest_Success(t *testing.T) {
	v := New()

	req := entity.VerifyOTPRequest{
		Token: "valid-session-token",
		Code:  "123456",
	}

	err := v.ValidateStruct(&req)
	assert.NoError(t, err)
}

func TestValidator_ValidateVerifyOTPRequest_MissingToken(t *testing.T) {
	v := New()

	req := entity.VerifyOTPRequest{
		Code: "123456",
	}

	err := v.ValidateStruct(&req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestValidator_ValidateVerifyOTPRequest_MissingCode(t *testing.T) {
	v := New()

	req := entity.VerifyOTPRequest{
		Token: "valid-session-token",
	}

	err := v.ValidateStruct(&req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code")
}

func TestValidator_FormatFieldError_PhoneNumberError(t *testing.T) {
	v := New()

	// Create a validation error for phone number
	req := entity.SendOTPRequest{PhoneNumber: "invalid"}
	err := v.ValidateStruct(&req)

	assert.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "phone_number")
	assert.Contains(t, errMsg, "must be a valid phone number")
}

func TestValidator_FormatFieldError_RequiredError(t *testing.T) {
	v := New()

	// Create a validation error for missing required field
	req := entity.SendOTPRequest{}
	err := v.ValidateStruct(&req)

	assert.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "phone_number")
	assert.Contains(t, errMsg, "is required")
}

func TestValidator_ValidateUser_Success(t *testing.T) {
	v := New()

	user := entity.User{
		PhoneNumber: "+1234567890",
	}

	err := v.ValidateStruct(&user)
	assert.NoError(t, err)
}

func TestValidator_ValidateUser_InvalidPhoneNumber(t *testing.T) {
	v := New()

	user := entity.User{
		PhoneNumber: "invalid-phone",
	}

	err := v.ValidateStruct(&user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone_number")
}

func TestValidator_ValidateStruct_NilInput(t *testing.T) {
	v := New()

	err := v.ValidateStruct(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestValidator_ValidateStruct_NonStruct(t *testing.T) {
	v := New()

	err := v.ValidateStruct("not a struct")
	assert.Error(t, err)
}

// Test the direct validatePhoneNumber function
func TestValidatePhoneNumber_Direct(t *testing.T) {
	// Create a validator instance to access the custom validation function
	v := validator.New()
	v.RegisterValidation("phone_number", validatePhoneNumber)

	validPhones := []string{
		"+1234567890",
		"+12345678901234",
		"+987654321098765",
	}

	for _, phone := range validPhones {
		err := v.Var(phone, "phone_number")
		assert.NoError(t, err, "Phone number %s should be valid", phone)
	}

	invalidPhones := []string{
		"1234567890",
		"+0234567890",
		"+12345",
		"salamsalam",
		"+abc1234567890",
	}

	for _, phone := range invalidPhones {
		err := v.Var(phone, "phone_number")
		assert.Error(t, err, "Phone number %s should be invalid", phone)
	}
}

func TestValidator_ComplexValidationScenarios(t *testing.T) {
	v := New()

	testCases := []struct {
		name        string
		phoneNumber string
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid US Number",
			phoneNumber: "+1234567890",
			expectError: false,
		},
		{
			name:        "Valid UK Number",
			phoneNumber: "+449876543210",
			expectError: false,
		},
		{
			name:        "Valid German Number",
			phoneNumber: "+491571234567",
			expectError: false,
		},
		{
			name:        "Valid Chinese Number",
			phoneNumber: "+8613912345678",
			expectError: false,
		},
		{
			name:        "Invalid - No Plus",
			phoneNumber: "1234567890",
			expectError: true,
			errorText:   "must be a valid phone number",
		},
		{
			name:        "Invalid - Starts with Zero",
			phoneNumber: "+0234567890",
			expectError: true,
			errorText:   "must be a valid phone number",
		},
		{
			name:        "Invalid - Too Short",
			phoneNumber: "+12345",
			expectError: true,
			errorText:   "must be a valid phone number",
		},
		{
			name:        "Invalid - Too Long",
			phoneNumber: "+123456789012345678",
			expectError: true,
			errorText:   "must be a valid phone number",
		},
		{
			name:        "Invalid - Contains Letters",
			phoneNumber: "+1234abc567890",
			expectError: true,
			errorText:   "must be a valid phone number",
		},
		{
			name:        "Invalid - Special Test Case",
			phoneNumber: "salamsalam",
			expectError: true,
			errorText:   "must be a valid phone number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := entity.SendOTPRequest{PhoneNumber: tc.phoneNumber}
			err := v.ValidateStruct(&req)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorText != "" {
					assert.Contains(t, err.Error(), tc.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
