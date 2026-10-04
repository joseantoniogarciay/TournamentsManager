package config

import "testing"

func appleEnvironment() map[string]string {
	return map[string]string{"APPLE_SERVICE_ID": "com.fasttourney.web.dev", "APPLE_TEAM_ID": "TEAM123456", "APPLE_KEY_ID": "KEY1234567", "APPLE_PRIVATE_KEY_FILE": "/tmp/example-apple-key.p8", "APPLE_REDIRECT_URI": "https://dev-api.fasttourney.com/v1/apple-callback", "APPLE_NATIVE_SCHEME": "fasttourney-dev"}
}

func TestAppleRequiresEveryRealConfigurationValue(t *testing.T) {
	values := appleEnvironment()
	read := func(key string) string { return values[key] }
	apple, err := loadApple(read)
	if err != nil || !apple.Enabled {
		t.Fatalf("configuration=%#v error=%v", apple, err)
	}
	for key, original := range values {
		for _, placeholder := range []string{"", "PLACEHOLDER", "REPLACE_WITH_VALUE", "replace-with-value", "<TEAM_ID>"} {
			values[key] = placeholder
			apple, err := loadApple(read)
			if err != nil || apple.Enabled {
				t.Fatalf("placeholder enabled %s", key)
			}
		}
		values[key] = original
	}
}

func TestAppleRejectsUnsafeCallbackAndNativeScheme(t *testing.T) {
	for _, callback := range []string{"http://api.example.test/v1/apple-callback", "https://user:password@api.example.test/v1/apple-callback", "https://api.example.test/other", "https://api.example.test/v1/apple-callback?redirect=evil", "https://api.example.test/v1/apple-callback#token"} {
		values := appleEnvironment()
		values["APPLE_REDIRECT_URI"] = callback
		if _, err := loadApple(func(key string) string { return values[key] }); err == nil {
			t.Fatalf("unsafe callback accepted: %s", callback)
		}
	}
	values := appleEnvironment()
	values["APPLE_NATIVE_SCHEME"] = "evil"
	if _, err := loadApple(func(key string) string { return values[key] }); err == nil {
		t.Fatal("arbitrary scheme accepted")
	}
}
