# Demo: DNT-Vault SSH Config Sync

## Scenario: Sync SSH config between work laptop and home desktop

### Step 1: Start the vault server (on VPS or local)

```bash
# On your VPS or a machine that's always accessible
export PORT=8443
export DATA_PATH=/var/lib/dnt-vault/data
export CONFIG_PATH=/etc/dnt-vault

./server/bin/dnt-vault-server
```

Output:
```
Creating default admin user...
Default user created: admin/admin
⚠️  Please change the default password!
DNT-Vault server starting on 0.0.0.0:8443
Data path: /var/lib/dnt-vault/data
Config path: /etc/dnt-vault
```

---

### Step 2: Setup on Work Laptop

```bash
# Initialize
./cli/bin/dnt-vault init
```

Interactive prompts:
```
Welcome to DNT-Vault SSH Config Sync!

Server Setup:
  Server URL [http://localhost:8443]: http://your-vps.com:8443

✓ Server configured: http://your-vps.com:8443

Master Password Setup:
This password encrypts your SSH configs.
  Enter master password: ****
  Confirm password: ****

✓ Master key generated and saved to ~/.dnt-vault/master.key
✓ Configuration saved to ~/.dnt-vault/config.yaml

Run 'dnt-vault login' to authenticate with the vault.
```

```bash
# Login
./cli/bin/dnt-vault login
```

Output:
```
Vault Server: http://your-vps.com:8443
Username: admin
Password: ****

✓ Logged in successfully
Token saved to ~/.dnt-vault/token
```

```bash
# Push your SSH config
./cli/bin/dnt-vault push --include-keys
```

Output:
```
Analyzing SSH config...
  Config file: ~/.ssh/config
  Hosts found: 5
  Referenced keys:
    - ~/.ssh/id_rsa
    - ~/.ssh/id_ed25519

Profile name [work-laptop]: 

⚠ You are about to upload private keys to the vault.
  Keys will be encrypted with a passphrase.

Continue? [y/N]: y

Enter passphrase for key encryption: ****
Confirm passphrase: ****

Encrypting config...
Encrypting keys (2 files)...
Uploading to vault...

✓ Profile 'work-laptop' pushed successfully
  Updated: 2026-05-04 10:30:00
  Config hash: abc123...
  Keys: 2 files uploaded
```

---

### Step 3: Pull on Home Desktop

```bash
# Initialize
./cli/bin/dnt-vault init
```

```
Server URL: http://your-vps.com:8443
Master password: **** (same as work laptop)
```

```bash
# Login
./cli/bin/dnt-vault login
```

```bash
# List available profiles
./cli/bin/dnt-vault list
```

Output:
```
Profiles on vault (http://your-vps.com:8443):

  work-laptop (work-laptop)
    Updated: 2 hours ago
    Keys: 2 files
    Hash: abc123...

Total: 1 profiles
```

```bash
# Pull the profile
./cli/bin/dnt-vault pull
```

Output:
```
Fetching profiles from vault...

Available profiles:
  1. work-laptop (work-laptop) - Updated: 2 hours ago - 2 keys

Select profile [1-1]: 1

Downloading profile 'work-laptop'...
Decrypting config...

⚠ Local SSH config exists and differs from vault.

Diff:
  --- Local
  +++ Remote (work-laptop)
  @@ -1,3 +1,8 @@
   Host example
       HostName example.com
  +    User ubuntu
  +
  +Host github
  +    HostName github.com
  +    User git

Abort pull? [Y/n]: n

Creating backup: ~/.dnt-vault/backups/2026-05-04_10-30-00.bak
Writing to ~/.ssh/config...

⚠ This profile includes 2 private keys.
  Decrypt and restore keys? [y/N]: y

Enter passphrase for key decryption: ****

Decrypting keys...
  ✓ id_rsa -> ~/.ssh/id_rsa
  ✓ id_ed25519 -> ~/.ssh/id_ed25519

✓ Profile 'work-laptop' pulled successfully
  Config restored
  2 keys restored
  Backup saved: ~/.dnt-vault/backups/2026-05-04_10-30-00.bak
```

---

### Step 4: Update and Re-sync

On work laptop, after adding a new host:

```bash
# Push updated config
./cli/bin/dnt-vault push
```

On home desktop:

```bash
# Pull latest changes
./cli/bin/dnt-vault pull --profile work-laptop
```

---

## Advanced Usage

### Multiple Profiles

```bash
# Push with custom name
./cli/bin/dnt-vault push --profile work-vpn

# Push another profile
./cli/bin/dnt-vault push --profile personal

# List all
./cli/bin/dnt-vault list
```

Output:
```
Profiles on vault:

  work-laptop (work-laptop)
    Updated: 2 hours ago
    Keys: 2 files

  work-vpn (work-laptop)
    Updated: 1 hour ago
    Keys: 1 file

  personal (home-desktop)
    Updated: 3 days ago
    Keys: none

Total: 3 profiles
```

### Delete Old Profiles

```bash
./cli/bin/dnt-vault delete --profile old-laptop
```

Output:
```
⚠ Delete profile 'old-laptop' from vault?
  This action cannot be undone.

Confirm [y/N]: y

Deleting profile...
✓ Profile 'old-laptop' deleted from vault
```

---

## Security Notes

1. **Master password**: Never shared with server, stored locally only
2. **Private keys**: Encrypted with separate passphrase before upload
3. **Transport**: Use HTTPS in production (setup TLS on server)
4. **Backups**: Automatic backups before each pull
5. **Tokens**: JWT tokens expire after 24 hours

---

## Production Deployment

### Server with TLS

```bash
# Get Let's Encrypt certificate
sudo certbot certonly --standalone -d vault.yourdomain.com

# Update server config
export PORT=443
export TLS_CERT=/etc/letsencrypt/live/vault.yourdomain.com/fullchain.pem
export TLS_KEY=/etc/letsencrypt/live/vault.yourdomain.com/privkey.pem

./server/bin/dnt-vault-server
```

### Client with HTTPS

```bash
# Update client config
vim ~/.dnt-vault/config.yaml
```

```yaml
server:
  url: https://vault.yourdomain.com
  tls_verify: true
```

---

## Troubleshooting

### Forgot master password?

Unfortunately, there's no recovery. The master password is used to derive the encryption key. Without it, encrypted data cannot be decrypted.

**Solution**: Delete `~/.dnt-vault/` and re-initialize. You'll need to push configs again.

### Forgot key passphrase?

Keys are encrypted separately. If you forget the passphrase, you can still pull the SSH config without keys.

**Solution**: Push again with `--include-keys` using a new passphrase.

### Server unreachable?

```bash
# Check server status
curl http://your-vps.com:8443/api/v1/profiles

# Check client config
cat ~/.dnt-vault/config.yaml
```
