package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateEmailCode 测试生成邮件验证码
func TestGenerateEmailCode(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"生成6位验证码"},
		{"生成6位验证码2"},
		{"生成6位验证码3"},
	}

	codes := make(map[string]bool)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := generateEmailCode()
			require.NoError(t, err)

			// 验证长度为6
			assert.Len(t, code, 6)

			// 验证全为数字
			for _, c := range code {
				assert.True(t, c >= '0' && c <= '9', "验证码应该只包含数字")
			}

			codes[code] = true
		})
	}

	// 验证多次调用生成不同的验证码
	assert.Greater(t, len(codes), 1, "多次调用应该生成不同的验证码")
}

// TestEmailCodeFormat 测试邮件验证码格式要求
func TestEmailCodeFormat(t *testing.T) {
	for i := 0; i < 10; i++ {
		code, err := generateEmailCode()
		require.NoError(t, err)

		// 验证长度为 6
		assert.Len(t, code, 6)

		// 验证全为数字且在 0-9 范围内
		for j, c := range code {
			assert.True(t, c >= '0' && c <= '9', "第 %d 个字符 %c 不是数字", j, c)
		}
	}
}

// TestEmailCodeRandomness 测试邮件验证码随机性
func TestEmailCodeRandomness(t *testing.T) {
	codes := make(map[string]int)

	// 生成 100 个验证码
	for i := 0; i < 100; i++ {
		code, err := generateEmailCode()
		require.NoError(t, err)
		codes[code]++
	}

	// 验证没有重复的验证码（或非常少）
	duplicates := 0
	for _, count := range codes {
		if count > 1 {
			duplicates += count - 1
		}
	}

	// 允许少量重复（概率上允许）
	assert.Less(t, duplicates, 5, "验证码重复过多：%d", duplicates)

	// 验证至少生成了大量不同的验证码
	assert.Greater(t, len(codes), 90, "应该生成足够多的不同验证码")
}
