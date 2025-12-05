# Security Quick Reference for Developers

## Safe Command Execution

### ✅ DO: Use Direct Execution
```go
// SECURE: No shell interpretation
cmd := exec.Command("npm", "run", "build", "--production")
cmd.Run()
```

### ❌ DON'T: Use Shell with User Input
```go
// VULNERABLE: Shell injection risk
userInput := getInput()
cmd := exec.Command("sh", "-c", "npm run "+userInput)  // DANGER!
```

---

## Input Validation

### Always Validate User Input
```go
const (
    MaxArgumentLength = 8192
    MaxArgumentCount  = 100
)

func validateArguments(args []string) error {
    if len(args) > MaxArgumentCount {
        return fmt.Errorf("too many arguments")
    }
    
    for i, arg := range args {
        if len(arg) > MaxArgumentLength {
            return fmt.Errorf("argument %d too long", i)
        }
    }
    
    return nil
}
```

---

## File Operations

### ✅ DO: Validate Paths
```go
func validateConfigPath(path string) error {
    if filepath.IsAbs(path) {
        return fmt.Errorf("absolute paths not allowed")
    }
    
    if strings.Contains(filepath.Clean(path), "..") {
        return fmt.Errorf("directory traversal not allowed")
    }
    
    return nil
}
```

### ❌ DON'T: Trust User-Provided Paths
```go
// VULNERABLE
func loadFile(userPath string) {
    data, _ := os.ReadFile(userPath)  // Path traversal risk!
}
```

---

## Resource Limits

### Always Set Limits
```go
const MaxConfigFileSize = 1024 * 1024  // 1MB

// Check file size before reading
info, _ := os.Stat(path)
if info.Size() > MaxConfigFileSize {
    return errors.New("file too large")
}
```

---

## Timeouts

### Use Context Timeouts
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

cmd := exec.CommandContext(ctx, program, args...)
if err := cmd.Run(); err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("command timed out")
    }
}
```

---

## Error Messages

### ✅ DO: Sanitize Error Messages
```go
// SAFE: Generic error
return fmt.Errorf("config file not accessible: %w", err)
```

### ❌ DON'T: Leak Sensitive Information
```go
// UNSAFE: Leaks file paths
return fmt.Errorf("failed to read %s: %w", absolutePath, err)
```

---

## Testing Security

### Write Security-Focused Tests
```go
func TestCommandInjection(t *testing.T) {
    injectionAttempts := []string{
        "; rm -rf /",
        "| cat /etc/passwd",
        "$(whoami)",
        "`id`",
        "&& malicious",
    }
    
    for _, attack := range injectionAttempts {
        // Verify injection is prevented
        err := executeCommand("echo", []string{attack})
        // Should NOT execute the injected command
    }
}
```

---

## Quick Checklist

Before committing code that handles user input:

- [ ] No shell interpretation (`sh -c`, `bash -c`)
- [ ] Input validation with limits
- [ ] Path validation (no absolute paths, no `..`)
- [ ] File size limits enforced
- [ ] Timeouts on operations
- [ ] Error messages don't leak paths/data
- [ ] Nil pointer checks
- [ ] Security tests written

---

## Common Pitfalls

### 1. String Concatenation for Commands
```go
// WRONG:
cmd := "npm run " + userInput
exec.Command("sh", "-c", cmd)

// RIGHT:
exec.Command("npm", "run", userInput)
```

### 2. Unchecked File Operations
```go
// WRONG:
data, _ := os.ReadFile(userPath)

// RIGHT:
if err := validatePath(userPath); err != nil {
    return err
}
info, err := os.Stat(userPath)
if err != nil || info.Size() > MaxSize {
    return err
}
data, err := os.ReadFile(userPath)
```

### 3. Missing Context Timeouts
```go
// WRONG:
cmd := exec.Command("long-running-process")

// RIGHT:
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
cmd := exec.CommandContext(ctx, "long-running-process")
```

---

## Resources

- **This Project's Security Audit**: See `SECURITY_AUDIT.md`
- **Go Security**: https://go.dev/doc/security/best-practices
- **OWASP**: https://owasp.org/www-project-go-secure-coding-practices-guide/

---

**Last Updated**: December 5, 2024
