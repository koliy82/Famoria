package config

import (
	"os"
	"strings"
	"testing"
)

func strp(s string) *string { return &s }

// caseSensitiveEnv is an in-memory environment with Linux semantics.
//
// Real environment variables are case-insensitive on Windows but case-sensitive
// on Linux. Testing promotion through os.Setenv would therefore pass trivially on
// Windows (both spellings are the same slot) and exercise the logic only on
// Linux, which is exactly the platform where the bug bites. An explicit map gives
// both platforms the same, correct behaviour.
type caseSensitiveEnv map[string]string

func (e caseSensitiveEnv) lookup(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

func (e caseSensitiveEnv) set(key, value string) { e[key] = value }

// TestPromoteKeysPromotesLowercase is the core regression guard.
//
// envconfig only ever looks up UPPER-CASE names, and godotenv preserves whatever
// spelling the .env file used. A lower-case "bot_proxy=true" in a Linux container
// would be silently ignored and the bot would connect direct — which on a network
// where Telegram and MongoDB are blocked shows up as an opaque timeout, not a
// configuration error.
func TestPromoteKeysPromotesLowercase(t *testing.T) {
	env := caseSensitiveEnv{
		"bot_proxy":     "true",
		"bot_proxy_url": "user:pass@1.2.3.4:8080",
		"db_proxy":      "true",
		"db_proxy_url":  "dbuser:dbpass@5.6.7.8:9090",
	}

	promoteKeys(lowercaseEnvKeys, env.lookup, env.set)

	want := map[string]string{
		"BOT_PROXY":     "true",
		"BOT_PROXY_URL": "user:pass@1.2.3.4:8080",
		"DB_PROXY":      "true",
		"DB_PROXY_URL":  "dbuser:dbpass@5.6.7.8:9090",
	}
	for k, v := range want {
		if got, ok := env.lookup(k); !ok || got != v {
			t.Errorf("%s = %q (present=%v), want %q", k, got, ok, v)
		}
	}
}

// TestPromoteKeysDoesNotOverrideExplicit verifies an explicitly set upper-case
// value wins. Without this, a lower-case leftover in .env would silently override
// an intentional environment setting.
func TestPromoteKeysDoesNotOverrideExplicit(t *testing.T) {
	env := caseSensitiveEnv{
		"BOT_PROXY": "false",
		"bot_proxy": "true",
	}

	promoteKeys(lowercaseEnvKeys, env.lookup, env.set)

	if got, _ := env.lookup("BOT_PROXY"); got != "false" {
		t.Errorf("BOT_PROXY = %q, want the explicit %q to win", got, "false")
	}
}

// TestPromoteKeysPreservesEmptyExplicitValue distinguishes "set to empty" from
// "not set": an explicitly empty upper-case value must not be replaced.
func TestPromoteKeysPreservesEmptyExplicitValue(t *testing.T) {
	env := caseSensitiveEnv{
		"BOT_PROXY_URL": "",
		"bot_proxy_url": "user:pass@1.2.3.4:8080",
	}

	promoteKeys(lowercaseEnvKeys, env.lookup, env.set)

	if got, _ := env.lookup("BOT_PROXY_URL"); got != "" {
		t.Errorf("BOT_PROXY_URL = %q, want it left empty", got)
	}
}

// TestPromoteKeysIgnoresUnlistedKeys ensures promotion is scoped to the known
// proxy keys and does not reshape the rest of the environment.
func TestPromoteKeysIgnoresUnlistedKeys(t *testing.T) {
	env := caseSensitiveEnv{
		"telegram_token": "abc",
		"mongo_uri":      "mongodb://x",
	}

	promoteKeys(lowercaseEnvKeys, env.lookup, env.set)

	if _, ok := env.lookup("TELEGRAM_TOKEN"); ok {
		t.Error("an unlisted key was promoted; only proxy keys should be touched")
	}
	if _, ok := env.lookup("MONGO_URI"); ok {
		t.Error("an unlisted key was promoted; only proxy keys should be touched")
	}
	if got, _ := env.lookup("telegram_token"); got != "abc" {
		t.Error("unlisted keys must be left untouched")
	}
}

// TestPromoteKeysHandlesMissingLowercase verifies an absent lower-case key is a
// no-op rather than an error or an empty promotion.
func TestPromoteKeysHandlesMissingLowercase(t *testing.T) {
	env := caseSensitiveEnv{}

	promoteKeys(lowercaseEnvKeys, env.lookup, env.set)

	if len(env) != 0 {
		t.Errorf("promoted %d keys from an empty environment", len(env))
	}
}

// TestNormalizeEnvKeysIsSafeToCall checks the real entry point does not panic on
// an empty environment, since it runs on every startup.
func TestNormalizeEnvKeysIsSafeToCall(t *testing.T) {
	normalizeEnvKeys()
}

// TestEveryListedKeyIsLowercase guards the table itself: a stray upper-case entry
// would make promoteKeys map a key onto itself, which is harmless but signals the
// list was edited without understanding its purpose.
func TestEveryListedKeyIsLowercase(t *testing.T) {
	if len(lowercaseEnvKeys) == 0 {
		t.Fatal("lowercaseEnvKeys is empty")
	}
	seen := make(map[string]bool, len(lowercaseEnvKeys))
	for _, k := range lowercaseEnvKeys {
		if k != strings.ToLower(k) {
			t.Errorf("key %q in lowercaseEnvKeys is not lower-case", k)
		}
		if seen[k] {
			t.Errorf("duplicate key %q in lowercaseEnvKeys", k)
		}
		seen[k] = true
	}
}

// TestValidateProxies covers the fail-fast behaviour: enabling a proxy without a
// usable URL must be an error, because the alternative is a silent direct
// connection that hangs on a blocked network.
func TestValidateProxies(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
		errHas  string
	}{
		{
			name: "both disabled",
			cfg:  Config{},
		},
		{
			name: "bot proxy configured",
			cfg: Config{
				BotProxy:    true,
				BotProxyURL: strp("user:pass@1.2.3.4:8080"),
			},
		},
		{
			name: "db proxy configured with https scheme",
			cfg: Config{
				DBProxy:    true,
				DBProxyURL: strp("https://user:pass@1.2.3.4:8443"),
			},
		},
		{
			name: "bot proxy enabled but URL missing",
			cfg: Config{
				BotProxy: true,
			},
			wantErr: true,
			errHas:  "BOT_PROXY",
		},
		{
			name: "bot proxy enabled but URL empty",
			cfg: Config{
				BotProxy:    true,
				BotProxyURL: strp("   "),
			},
			wantErr: true,
			errHas:  "BOT_PROXY",
		},
		{
			name: "db proxy enabled but URL malformed",
			cfg: Config{
				DBProxy:    true,
				DBProxyURL: strp("not a url"),
			},
			wantErr: true,
			errHas:  "DB_PROXY",
		},
		{
			name: "db proxy enabled but URL missing",
			cfg: Config{
				DBProxy: true,
			},
			wantErr: true,
			errHas:  "DB_PROXY",
		},
		{
			name: "url set while proxy disabled is allowed",
			cfg: Config{
				BotProxyURL: strp("user:pass@1.2.3.4:8080"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cfg.validateProxies()
			if c.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), c.errHas) {
					t.Errorf("error %q should mention %s", err, c.errHas)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestValidateProxiesErrorDoesNotLeakPassword ensures a malformed URL reported at
// startup does not print the credentials it was given.
func TestValidateProxiesErrorDoesNotLeakPassword(t *testing.T) {
	const secret = "supersecret"
	cfg := Config{
		BotProxy:    true,
		BotProxyURL: strp("https://user:" + secret + "@1.2.3.4"), // no port
	}

	err := cfg.validateProxies()
	if err == nil {
		t.Fatal("expected an error for a proxy URL with no port")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("validation error leaked the password: %v", err)
	}
}

// TestNewReadsLowercaseEnvFile is the end-to-end check: a real .env file on disk
// written with lower-case proxy keys must produce a validated Config.
//
// Note that on Windows the process environment is itself case-insensitive, so
// this passes there even without normalization; the deterministic coverage of
// the promotion logic is TestPromoteKeysPromotesLowercase, which uses an
// explicitly case-sensitive map. This test covers the rest of the chain —
// godotenv parsing, promotion, envconfig decoding and validation.
func TestNewReadsLowercaseEnvFile(t *testing.T) {
	// New() reads .env from the working directory, so run it in a temp one.
	dir := t.TempDir()
	envFile := dir + string(os.PathSeparator) + ".env"

	content := strings.Join([]string{
		"TELEGRAM_TOKEN=test-token",
		"MONGO_URI=mongodb://localhost:27017",
		"MONGO_DATABASE=testdb",
		"APP_TIMEZONE=Europe/Moscow",
		// Lower case, exactly as the task specifies.
		"bot_proxy=true",
		"bot_proxy_url=botuser:botpass@1.2.3.4:8080",
		"db_proxy=true",
		"db_proxy_url=dbuser:dbpass@5.6.7.8:9090",
	}, "\n") + "\n"

	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// godotenv.Load does not overwrite variables that are already set, so every
	// relevant key must be cleared or a value inherited from the developer's real
	// .env would mask the file under test. Snapshot and restore afterwards, since
	// t.Setenv cannot express "unset".
	keys := []string{
		"TELEGRAM_TOKEN", "MONGO_URI", "MONGO_DATABASE", "APP_TIMEZONE",
		"bot_proxy", "BOT_PROXY", "bot_proxy_url", "BOT_PROXY_URL",
		"db_proxy", "DB_PROXY", "db_proxy_url", "DB_PROXY_URL",
		"YTDLP_COOKIES_FILE", "PROXY_ENABLE", "PROXY_URL",
	}
	saved := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			saved[k] = v
		}
		_ = os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for _, k := range keys {
			_ = os.Unsetenv(k)
		}
		for k, v := range saved {
			_ = os.Setenv(k, v)
		}
	})

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	cfg := New()

	if !cfg.BotProxy {
		t.Error("BotProxy = false, want true from lower-case bot_proxy")
	}
	if cfg.BotProxyURL == nil || *cfg.BotProxyURL != "botuser:botpass@1.2.3.4:8080" {
		t.Errorf("BotProxyURL = %v, want the value from bot_proxy_url", cfg.BotProxyURL)
	}
	if !cfg.DBProxy {
		t.Error("DBProxy = false, want true from lower-case db_proxy")
	}
	if cfg.DBProxyURL == nil || *cfg.DBProxyURL != "dbuser:dbpass@5.6.7.8:9090" {
		t.Errorf("DBProxyURL = %v, want the value from db_proxy_url", cfg.DBProxyURL)
	}
	if err := cfg.validateProxies(); err != nil {
		t.Errorf("validateProxies on a fully configured pair: %v", err)
	}
}

// TestNewRejectsEnabledProxyWithoutURL asserts the fail-fast path through the
// real constructor, not just the validator in isolation.
func TestNewRejectsEnabledProxyWithoutURL(t *testing.T) {
	dir := t.TempDir()
	envFile := dir + string(os.PathSeparator) + ".env"

	content := strings.Join([]string{
		"TELEGRAM_TOKEN=test-token",
		"MONGO_URI=mongodb://localhost:27017",
		"MONGO_DATABASE=testdb",
		"APP_TIMEZONE=Europe/Moscow",
		"bot_proxy=true",
		// bot_proxy_url deliberately absent
	}, "\n") + "\n"
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	keys := []string{
		"TELEGRAM_TOKEN", "MONGO_URI", "MONGO_DATABASE", "APP_TIMEZONE",
		"bot_proxy", "BOT_PROXY", "bot_proxy_url", "BOT_PROXY_URL",
		"db_proxy", "DB_PROXY", "db_proxy_url", "DB_PROXY_URL",
	}
	saved := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			saved[k] = v
		}
		_ = os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for _, k := range keys {
			_ = os.Unsetenv(k)
		}
		for k, v := range saved {
			_ = os.Setenv(k, v)
		}
	})

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("New() should panic when BOT_PROXY is true and BOT_PROXY_URL is empty")
		}
		msg, ok := r.(error)
		if !ok {
			if s, isStr := r.(string); isStr {
				t.Logf("panic value (string): %s", s)
				return
			}
			t.Fatalf("panic value has unexpected type %T", r)
		}
		if !strings.Contains(msg.Error(), "BOT_PROXY") {
			t.Errorf("panic %v should name BOT_PROXY", msg)
		}
	}()

	New()
}
