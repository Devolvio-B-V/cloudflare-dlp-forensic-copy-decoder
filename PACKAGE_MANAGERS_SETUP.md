# Package Manager Publishing Setup Guide

This guide explains how to set up automated publishing to Homebrew and Winget package managers.

## Prerequisites

- Admin access to the Devolvio-B-V GitHub organization
- Ability to create repositories and secrets

## 1. Homebrew Setup (macOS/Linux)

### One-Time Setup

1. **Create the Homebrew tap repository**
   - Go to https://github.com/organizations/Devolvio-B-V/repositories/new
   - Repository name: `homebrew-tap`
   - Description: "Homebrew formulae for Devolvio B.V. tools"
   - Visibility: Public
   - Click "Create repository"
   - **Note**: Leave the repository empty - GoReleaser will populate it automatically

2. **Generate a Personal Access Token (PAT)**
   - Go to https://github.com/settings/tokens/new
   - Token name: `GoReleaser Homebrew Tap`
   - Expiration: No expiration (or choose your preference)
   - Select scopes:
     - ✅ `repo` (Full control of private repositories)
   - Click "Generate token"
   - **Important**: Copy the token immediately (you won't see it again)

3. **Add the token as a repository secret**
   - Go to https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/settings/secrets/actions
   - Click "New repository secret"
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Value: Paste the token from step 2
   - Click "Add secret"

### How It Works

When you create a release tag (e.g., `v2.0.0`):
1. GitHub Actions triggers the release workflow
2. GoReleaser builds the binaries
3. GoReleaser automatically creates/updates the formula in `homebrew-tap`
4. Users can install with: `brew install devolvio-b-v/tap/cf-dlp-decode`

## 2. Winget Setup (Windows)

Winget publishing is **fully automated** via GitHub Actions. The workflow automatically creates PRs to microsoft/winget-pkgs on each release.

### One-Time Setup

1. **Fork the winget-pkgs repository**
   - Go to https://github.com/microsoft/winget-pkgs
   - Click "Fork" in the top-right corner
   - This creates a fork at: `https://github.com/YOUR_USERNAME/winget-pkgs`

2. **Generate a Personal Access Token (PAT)**
   - Go to https://github.com/settings/tokens/new
   - Token name: `Winget Publishing`
   - Expiration: No expiration (or choose your preference)
   - Select scopes:
     - ✅ `repo` (Full control of private repositories)
     - ✅ `public_repo` (Access public repositories)
   - Click "Generate token"
   - **Important**: Copy the token immediately

3. **Add the token as a repository secret**
   - Go to https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/settings/secrets/actions
   - Click "New repository secret"
   - Name: `WINGET_GITHUB_TOKEN`
   - Value: Paste the token from step 2
   - Click "Add secret"

### How It Works

When you create a release tag (e.g., `v2.0.0`):
1. GitHub Actions triggers the release workflow
2. GoReleaser builds and publishes the binaries
3. The `publish-winget` job automatically:
   - Downloads wingetcreate
   - Creates/updates the Winget manifest for both x64 and arm64 architectures
   - Submits a PR to microsoft/winget-pkgs via your fork
4. Microsoft reviews the PR (typically 1-3 business days)
5. Once merged, users can install with: `winget install Devolvio-B-V.cf-dlp-decode`

### Manual Submission (Fallback)

If the automated workflow fails, you can manually submit:

1. **Install wingetcreate**
   ```powershell
   winget install wingetcreate
   ```

2. **For each release**, run:
   ```powershell
   # Replace VERSION with your release version (e.g., 2.0.0)
   $VERSION = "2.0.0"
   
   # Generate manifest and submit PR
   wingetcreate update Devolvio-B-V.cf-dlp-decode `
     --version $VERSION `
     --urls "https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases/download/v$VERSION/cf-dlp-decode-windows-amd64.exe|x64" `
            "https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases/download/v$VERSION/cf-dlp-decode-windows-arm64.exe|arm64" `
     --token YOUR_GITHUB_TOKEN `
     --submit
   ```

3. **Wait for Microsoft review**
   - Microsoft team reviews the PR (typically 1-3 business days)
   - They may request changes or automatically merge

## Verification

After setting up:

1. **Create a test release**
   ```bash
   git tag v2.0.0-test
   git push origin v2.0.0-test
   ```

2. **Check GitHub Actions**
   - Go to https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/actions
   - Verify the "Release" workflow runs successfully
   - Check that all jobs complete: `release`, `publish-brew`, and `publish-winget`

3. **Verify package manager repositories**
   - Homebrew: Check https://github.com/Devolvio-B-V/homebrew-tap for the new formula
   - Winget: Check your microsoft/winget-pkgs fork for a new branch and PR

## Troubleshooting

### Homebrew: "Token authentication failed"
- Verify the PAT has the correct permissions (`repo` scope)
- Ensure the token hasn't expired
- Check the secret name matches exactly: `HOMEBREW_TAP_GITHUB_TOKEN`

### Homebrew: Repository not found
- Ensure the repository exists and is public
- Verify the repository name matches exactly: `homebrew-tap`

### Winget: PR submission failed
- Verify `WINGET_GITHUB_TOKEN` is set correctly in repository secrets
- Ensure your fork of microsoft/winget-pkgs exists
- Check that the token has `repo` and `public_repo` scopes
- Verify release artifacts are accessible (URLs return 200 OK)

### Winget: PR rejected by Microsoft
- Ensure all required manifest files are present
- Verify URLs are correct and accessible
- Check that version numbers match across all files
- Review Microsoft's feedback on the PR

## Additional Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Winget Package Submission](https://github.com/microsoft/winget-pkgs/blob/master/AUTHORING_MANIFESTS.md)
- [wingetcreate Documentation](https://github.com/microsoft/winget-create)
