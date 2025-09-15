package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildStartMessage(t *testing.T) {
	t.Parallel()

	expected := "👋 \\*Привет\\!\\* Отправь мне одноразовый код для привязки аккаунта, начинающийся с \\`auth\\_XXXX\\`\\.\n\nИли воспользуйся кнопкой ниже ⬇️"
	result := buildStartMessage()

	assert.Equal(t, expected, result)
	assert.Contains(t, result, "Привет")
	assert.Contains(t, result, "auth\\_XXXX")
	assert.True(t, strings.HasPrefix(result, "👋"))
}

func TestBuildSuccessMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple message",
			input:    "Успешно!",
			expected: "✅ Успешно\\!",
		},
		{
			name:     "message with markdown",
			input:    "Ваш аккаунт *привязан*",
			expected: "✅ Ваш аккаунт \\*привязан\\*",
		},
		{
			name:     "message with underscores",
			input:    "User_name connected",
			expected: "✅ User\\_name connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := buildSuccessMessage(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.True(t, strings.HasPrefix(result, "✅"))
		})
	}
}

func TestBuildErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple error",
			input:    "Ошибка!",
			expected: "❌ Ошибка\\!",
		},
		{
			name:     "error with special chars",
			input:    "Error (code: 404)",
			expected: "❌ Error \\(code: 404\\)",
		},
		{
			name:     "error with brackets",
			input:    "Invalid [input]",
			expected: "❌ Invalid \\[input\\]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := buildErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.True(t, strings.HasPrefix(result, "❌"))
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "underscore",
			input:    "test_underscore",
			expected: "test\\_underscore",
		},
		{
			name:     "asterisk",
			input:    "text*text",
			expected: "text\\*text",
		},
		{
			name:     "brackets",
			input:    "[text]",
			expected: "\\[text\\]",
		},
		{
			name:     "parentheses",
			input:    "(text)",
			expected: "\\(text\\)",
		},
		{
			name:     "multiple special chars",
			input:    "Hello *world* [test] (example)",
			expected: "Hello \\*world\\* \\[test\\] \\(example\\)",
		},
		{
			name:     "all special chars",
			input:    "_*[]()~`>#+-=|{}.!",
			expected: "\\_\\*\\[\\]\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!",
		},
		{
			name:     "colon not escaped",
			input:    "text: with colon",
			expected: "text: with colon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := escapeMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultKeyboard(t *testing.T) {
	t.Parallel()

	keyboard := defaultKeyboard()

	assert.Len(t, keyboard.Keyboard, 1, "Keyboard should have one row")
	assert.Len(t, keyboard.Keyboard[0], 1, "First row should have one button")

	button := keyboard.Keyboard[0][0]
	assert.Equal(t, "/packages", button.Text, "Button text should be /packages")
}
