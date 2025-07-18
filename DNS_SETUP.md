# DNS Configuration for gismo.run (Cloudflare)

To complete the setup of gismo.run for GitHub Pages, you need to configure your DNS settings in Cloudflare.

## Cloudflare DNS Configuration

### Important Cloudflare Settings:
- **Proxy status**: Set to "DNS only" (gray cloud) for GitHub Pages
- **SSL/TLS**: Set to "Full" in Cloudflare SSL/TLS settings

## Required DNS Records

Add the following DNS records in your Cloudflare dashboard:

### Option 1: Using APEX domain (Recommended)

Add these A records for the root domain:
```dns
Type: A
Name: @ (or gismo.run)
Value: 185.199.108.153
Proxy status: DNS only (gray cloud)
TTL: Auto

Type: A
Name: @ (or gismo.run)
Value: 185.199.109.153
Proxy status: DNS only (gray cloud)
TTL: Auto

Type: A
Name: @ (or gismo.run)
Value: 185.199.110.153
Proxy status: DNS only (gray cloud)
TTL: Auto

Type: A
Name: @ (or gismo.run)
Value: 185.199.111.153
Proxy status: DNS only (gray cloud)
TTL: Auto
```

### Option 2: Using www subdomain

If you prefer to use www.gismo.run:
```dns
Type: CNAME
Name: www
Value: jrossi.github.io
TTL: 3600 (or default)
```

Then add a redirect from gismo.run to www.gismo.run.

## Verification Steps

1. **Wait for DNS propagation** (usually 5-30 minutes, can take up to 48 hours)

2. **Check DNS propagation:**
   ```bash
   dig gismo.run +noall +answer
   # or
   nslookup gismo.run
   ```

3. **Enable GitHub Pages in repository settings:**
  - Go to https://github.com/jrossi/gismo/settings/pages
  - Under "Source", select "Deploy from a branch"
  - Select "main" branch and "/docs" folder
  - Click Save

4. **Verify custom domain:**
  - After DNS propagates, GitHub will automatically verify your domain
  - You should see a green checkmark next to your custom domain in the Pages settings

5. **Enable HTTPS (recommended):**
  - Once the domain is verified, check "Enforce HTTPS" in the Pages settings
  - This may take a few minutes to provision the SSL certificate

## Troubleshooting

- If you see a 404 error, ensure the docs workflow has run and deployed
- If DNS isn't resolving, double-check your DNS records
- GitHub's IP addresses for Pages: 185.199.108-111.153

## Current Status

- [x] CNAME files added to repository
- [ ] DNS records configured at registrar
- [ ] GitHub Pages enabled in repository settings
- [ ] Domain verified by GitHub
- [ ] HTTPS enabled