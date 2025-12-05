# Path Validation Security - Design Rationale

## Question: Why block absolute paths if users can just `cd` first?

### Short Answer
**Defense in depth** - Even non-perfect protections provide value.

### Long Answer

#### What This DOES Protect Against:

1. **Accidental Mistakes**
   ```bash
   # Typo or wrong path:
   tb build --config /old-project/config.yaml
   # ❌ Fails with clear error instead of loading wrong config
   ```

2. **Automated Attacks**
   ```bash
   # Malicious script:
   for file in /etc/*; do
       tb build --config $file  # ❌ All fail
   done
   ```

3. **Social Engineering**
   ```bash
   # Hidden in a complex command:
   tb build --very --many --flags --config /attacker/evil.yaml
   # ❌ Fails, makes attack obvious
   ```

4. **Best Practice Enforcement**
   - Encourages project-local configs
   - Makes scripts portable
   - Self-documenting (config in current dir)

#### What This DOES NOT Protect Against:

1. **Determined Users**
   ```bash
   cd /etc && tb build --config passwd  # ✅ Works
   ```

2. **Environment Compromise**
   - If attacker controls shell, path validation doesn't matter
   - If system is compromised, CLI tool security is irrelevant

3. **Root/Admin Attacks**
   - Users with admin rights can bypass anything

### The Real Security Model

For **local CLI tools**, the threat model is:

```
┌─────────────────────────────────────────┐
│ Threat: Accidental Misuse               │
│ Risk: HIGH (typos, mistakes)            │
│ Protection: Path validation ✅          │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Threat: Automated Script Attacks        │
│ Risk: MEDIUM (malware, bots)            │
│ Protection: Path validation ✅          │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Threat: Social Engineering              │
│ Risk: MEDIUM (hidden flags)             │
│ Protection: Path validation ✅          │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Threat: Determined Malicious User       │
│ Risk: LOW (local access)                │
│ Protection: Path validation ⚠️ (limited) │
│ Note: User can cd to any directory      │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Threat: Compromised Environment         │
│ Risk: CRITICAL (game over)              │
│ Protection: None (out of scope)         │
│ Note: OS-level security required        │
└─────────────────────────────────────────┘
```

### Defense in Depth Philosophy

Even "bypassable" protections are valuable because:

1. **Layers of Security**
   - Each barrier increases attack difficulty
   - Defense in depth: multiple controls working together

2. **Making Attacks Obvious**
   ```bash
   # Hidden attack (fails):
   malicious_script.sh --config /etc/shadow
   
   # Obvious attack (works but visible):
   cd /etc && malicious_script.sh --config shadow
   ```

3. **Preventing Accidents ≠ Preventing Attacks**
   - Primary goal: prevent user mistakes
   - Secondary benefit: makes some attacks harder

4. **Failing Safe**
   - When something goes wrong, fail with clear error
   - Better than silently loading wrong config

### Comparison to Web Applications

| CLI Tool | Web Application |
|----------|-----------------|
| User controls environment | No control over environment |
| Local file access | Remote file access |
| cd can bypass restrictions | Cannot cd |
| Primary threat: Mistakes | Primary threat: Attacks |
| Path validation: Helpful | Path validation: Critical |

### Conclusion

**Keep the absolute path prevention** because:

✅ Prevents common mistakes (primary goal)  
✅ Makes automated attacks harder  
✅ Enforces best practices  
✅ Provides clear user feedback  
✅ Defense in depth (even if not perfect)  

⚠️ **Accept that:**
- It's not a hard security boundary
- Users can work around it with `cd`
- That's okay - it still provides value

### Alternative Implementations

If you want **stricter security**:

```go
// Option 1: Allowlist only
allowedPaths := []string{
    ".",           // Current directory
    "~/.toolbox/", // User config
}

// Option 2: Confirmation for unusual paths
if isUnusualPath(path) {
    confirmWithUser("Load config from unusual location?")
}

// Option 3: Digital signatures (paranoid mode)
if !verifySignature(configFile) {
    return errors.New("unsigned config from untrusted location")
}
```

---

**Bottom Line**: Path validation is about **preventing mistakes** and **raising the bar for attacks**, not creating an impenetrable barrier. For a CLI tool, that's exactly the right goal.
