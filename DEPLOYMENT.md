# Deployment Guide

This guide explains how to deploy the Poker application to production using Render.com (free tier, no credit card required).

## Overview

The application is ready for deployment with:
- **Multi-stage Docker build** that builds both frontend and backend
- **Production-optimized image** (only 16.2MB)
- **Health checks** configured
- **Environment variable support**
- **WebSocket support** for real-time gameplay

## Prerequisites

1. GitHub account
2. Render.com account (sign up at https://render.com)
3. Your poker repository pushed to GitHub

## Deployment Options

### Option A: Deploy via Render.com Dashboard (Recommended)

#### Step 1: Prepare Your Repository

Ensure your code is pushed to GitHub:

```bash
git add .
git commit -m "Add production deployment configuration"
git push origin main
```

#### Step 2: Sign Up for Render

1. Go to https://render.com
2. Click "Get Started for Free"
3. Sign up with your GitHub account (no credit card required)
4. Authorize Render to access your GitHub repositories

#### Step 3: Create a New Web Service

1. From the Render Dashboard, click "New +" → "Web Service"
2. Select "Build and deploy from a Git repository"
3. Connect your poker repository
4. Click "Connect" next to your repository

#### Step 4: Configure the Service

Fill in the following settings:

- **Name**: `poker-app` (or your preferred name)
- **Region**: Choose closest to you (e.g., Frankfurt, Oregon, Singapore)
- **Branch**: `main`
- **Runtime**: `Docker`
- **Dockerfile Path**: `./Dockerfile.production`
- **Instance Type**: `Free`

#### Step 5: Environment Variables (Optional)

The defaults work fine, but you can customize:

- `PORT`: `8080` (automatically set by Render)
- `LOG_LEVEL`: `info` (options: debug, info, warn, error)

#### Step 6: Advanced Settings

- **Health Check Path**: `/health`
- **Auto-Deploy**: `Yes` (deploys automatically on git push)

#### Step 7: Deploy

1. Click "Create Web Service"
2. Render will:
   - Clone your repository
   - Build the Docker image (takes 3-5 minutes first time)
   - Deploy the container
   - Assign a public URL

3. Monitor the build logs in real-time
4. Once deployed, your app will be available at: `https://poker-app-xxxx.onrender.com`

### Option B: Deploy via render.yaml Blueprint (Advanced)

If you prefer infrastructure-as-code:

1. The `render.yaml` file is already configured in your repository
2. From Render Dashboard, click "New +" → "Blueprint"
3. Select your poker repository
4. Render will automatically detect and use `render.yaml`
5. Click "Apply" to deploy

## Accessing Your Deployed Application

### Frontend
- URL: `https://your-app-name.onrender.com`
- Opens the React poker interface

### WebSocket
- URL: `wss://your-app-name.onrender.com/ws`
- Used by the frontend for real-time gameplay

### Health Check
- URL: `https://your-app-name.onrender.com/health`
- Returns: `{"status":"ok"}`

## Free Tier Limitations

Render's free tier includes:

- ✅ **750 hours/month** of runtime (enough for one 24/7 service)
- ✅ **Automatic HTTPS** with SSL certificates
- ✅ **Custom domains** supported
- ✅ **Auto-deploy** on git push
- ⚠️ **Sleeps after 15 minutes** of inactivity
- ⚠️ **Cold start delay** (~30 seconds when waking up)
- ⚠️ **Shared CPU** and 512MB RAM

### How Sleep Works

- App goes to sleep after 15 minutes without requests
- First request after sleep takes ~30 seconds to wake up
- Subsequent requests are instant
- Perfect for demos, testing, and low-traffic apps

To keep it awake, you can:
- Use an uptime monitoring service (UptimeRobot, etc.)
- Upgrade to paid tier ($7/month for always-on)

## Post-Deployment Configuration

### Update WebSocket URL in Frontend (if needed)

The frontend should auto-detect the WebSocket URL, but if you need to configure it:

```typescript
// frontend/src/services/WebSocketService.ts
const WS_URL = process.env.NODE_ENV === 'production'
  ? `wss://${window.location.host}/ws`
  : 'ws://localhost:8080/ws'
```

### Configure CORS/WebSocket Origins (Production Security)

For production, you should restrict WebSocket origins in `internal/server/server.go`:

```go
upgrader: &websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        // Replace with your actual domain
        return origin == "https://poker-app-xxxx.onrender.com"
    },
},
```

## Continuous Deployment

Once deployed, Render automatically redeploys when you push to your main branch:

```bash
# Make changes
git add .
git commit -m "Add new feature"
git push origin main

# Render automatically:
# 1. Detects the push
# 2. Rebuilds the Docker image
# 3. Deploys the new version
# 4. Zero-downtime deployment
```

## Monitoring and Logs

### View Logs

1. Go to your service in Render Dashboard
2. Click "Logs" tab
3. See real-time structured JSON logs

### Metrics

Render provides:
- CPU usage
- Memory usage
- Request volume
- Response times
- Build history

### Health Checks

Render automatically monitors `/health` endpoint:
- Checks every 30 seconds
- Restarts service if unhealthy
- Alerts you via email

## Troubleshooting

### Build Fails

**Check the build logs** in Render Dashboard:

Common issues:
- Missing dependencies: Ensure `go.mod`, `go.sum`, and `package.json` are committed
- Docker build errors: Test locally with `docker build -f Dockerfile.production -t poker:test .`

### Service Won't Start

**Check runtime logs**:

Common issues:
- Port binding: Render sets `PORT` env var, ensure your app uses it
- Health check failing: Ensure `/health` endpoint is accessible
- Static files missing: Verify `web/static/` contains frontend build

### WebSocket Connection Fails

**Check browser console**:

Common issues:
- HTTPS/WSS mismatch: Use `wss://` for HTTPS sites, not `ws://`
- CORS/Origin issues: Check `CheckOrigin` function in `server.go`
- Firewall: Render should allow WebSocket by default

### App Sleeps Too Often

**Solutions**:
- Use uptime monitoring (free tier of UptimeRobot)
- Upgrade to paid plan ($7/month for always-on)
- Accept the tradeoff (good for demos)

## Scaling and Upgrades

### When to Upgrade

Consider upgrading from free tier when:
- You need 24/7 availability (no sleep)
- Traffic exceeds 100 concurrent users
- You need more RAM (>512MB)
- You want faster response times

### Render Paid Plans

- **Starter**: $7/month - Always-on, 512MB RAM
- **Standard**: $25/month - 2GB RAM, auto-scaling
- **Pro**: $85/month - 4GB RAM, dedicated CPU

## Alternative Deployment Platforms

If Render doesn't work for you:

### Railway.app
- Similar to Render
- $5 free credit/month
- May require credit card
- WebSocket-friendly

### Fly.io
- Edge deployment (global CDN)
- Free tier available
- Requires credit card
- Excellent WebSocket support

### Koyeb
- Free tier available (check current status)
- European-based
- Docker support

## Local Production Testing

Test the production build locally before deploying:

```bash
# Build the production image
docker build -f Dockerfile.production -t poker:production .

# Run it locally
docker run -p 8080:8080 poker:production

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/

# Stop
docker stop <container-id>
```

## Support

If you encounter issues:

1. Check Render's status page: https://status.render.com
2. Review Render documentation: https://docs.render.com
3. Check application logs in Render Dashboard
4. Test locally first with Docker
5. Open an issue on GitHub

## Summary Checklist

Before deploying:

- [ ] Code pushed to GitHub
- [ ] `Dockerfile.production` exists and tested
- [ ] `render.yaml` configured (optional)
- [ ] Environment variables reviewed
- [ ] Health endpoint works (`/health`)
- [ ] Frontend builds successfully
- [ ] WebSocket functionality tested

After deploying:

- [ ] Service is running (green status)
- [ ] Health check is passing
- [ ] Frontend loads in browser
- [ ] WebSocket connects successfully
- [ ] Gameplay works end-to-end
- [ ] Logs are clean (no errors)
- [ ] Custom domain configured (optional)

## Next Steps

1. Deploy to Render.com following the steps above
2. Test the deployed application thoroughly
3. Share the URL with others to play
4. Monitor logs and performance
5. Consider adding:
   - Database (PostgreSQL on Render)
   - Redis for session storage
   - Custom domain
   - Analytics

---

**Ready to deploy? Go to https://render.com and get started!**
