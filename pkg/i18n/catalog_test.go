// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package i18n

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

func TestResolveLocale(t *testing.T) {
	for _, tc := range []struct{ cookie, header, want string }{
		{"en", "zh-CN", "en"},
		{"zh-CN", "en", "zh-CN"},
		{"invalid", "en;q=0.4,zh-CN;q=0.9", "zh-CN"},
		{"", "ru-RU,en-US;q=0.8", "en"},
		{"", "zh-TW", "zh-CN"},
		{"", "zh;q=0,en;q=0.5", "en"},
		{"", "zh;q=invalid", "en"},
		{"zh-invalid-prefix", "", "zh-CN"},
		{"zhbad", "", "en"},
		{"", "", "en"},
	} {
		t.Run(tc.cookie+"/"+tc.header, func(t *testing.T) {
			if got := ResolveLocale(tc.cookie, tc.header); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTranslatePreservesArgumentsAndUnknownText(t *testing.T) {
	for _, tc := range []struct{ locale, source, want string }{
		{"zh-CN", "Scanning certificates due for renewal", "开始扫描待续期证书"},
		{"en", "开始扫描待续期证书", "Scanning certificates due for renewal"},
		{"en", "目标日志库不合法: 用户数据", "invalid target log database: 用户数据"},
		{"zh-CN", "invalid target log database: 用户数据", "目标日志库不合法: 用户数据"},
		{"en", "自定义名称，不是系统消息", "自定义名称，不是系统消息"},
		{"zh-CN", "An unknown third-party error", "An unknown third-party error"},
		{"unknown", "开始扫描待续期证书", "Scanning certificates due for renewal"},
		{"en", "smtp_invalid:该账号未绑定邮箱，使用临时码登录", "smtp_invalid:This account has no email address; use a temporary code to sign in"},
		{"zh-CN", "smtp_invalid:This account has no email address; use a temporary code to sign in", "smtp_invalid:该账号未绑定邮箱，使用临时码登录"},
		{"zh-CN", "need_email_code:60", "need_email_code:60"},
	} {
		t.Run(tc.locale+"/"+tc.source, func(t *testing.T) {
			if got := Translate(tc.locale, tc.source); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
	legacy := "[12:34:56] 开始扫描待续期证书\n[12:34:57] 自定义日志\n"
	want := "[12:34:56] Scanning certificates due for renewal\n[12:34:57] 自定义日志\n"
	if got := TranslateLog("en", legacy); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTranslateRecognizesHistoricalEnglishErrorCase(t *testing.T) {
	for _, source := range []string{
		"invalid target log database: custom-value",
		"Invalid target log database: custom-value",
	} {
		if got, want := Translate("zh-CN", source), "目标日志库不合法: custom-value"; got != want {
			t.Errorf("Translate(%q, %q) = %q, want %q", "zh-CN", source, got, want)
		}
	}
}

func TestCatalogParityAndFormats(t *testing.T) {
	variables := regexp.MustCompile(`\{\{[^{}]+\}\}`)
	english := builtIn.locales[DefaultLocale]
	for locale, messages := range builtIn.locales {
		if len(messages) != len(english) {
			t.Fatalf("%s has %d messages, want %d", locale, len(messages), len(english))
		}
		for id, source := range english {
			translated, ok := messages[id]
			if !ok || translated == "" {
				t.Errorf("%s is missing %s", locale, id)
			}
			if !reflect.DeepEqual(printfVerb.FindAllString(source, -1), printfVerb.FindAllString(translated, -1)) {
				t.Errorf("%s %s changes formatting verbs", locale, id)
			}
			sourceVariables := variables.FindAllString(source, -1)
			translatedVariables := variables.FindAllString(translated, -1)
			sort.Strings(sourceVariables)
			sort.Strings(translatedVariables)
			if !reflect.DeepEqual(sourceVariables, translatedVariables) {
				t.Errorf("%s %s changes template variables", locale, id)
			}
		}
	}
}

func TestTranslateLegacyTaskLog(t *testing.T) {
	source := "[系统] 任务已成功入队，等待调度执行 (队列: %s, 最大重试次数: %d)"
	got := Translate("en", fmt.Sprintf(source, "custom-中文", 3))
	want := "[System] Task enqueued; waiting for execution (queue: custom-中文, maximum retries: 3)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
