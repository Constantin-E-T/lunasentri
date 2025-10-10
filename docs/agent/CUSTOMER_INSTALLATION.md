# Install LunaSentri Agent - Customer Guide

Monitor your server in 60 seconds! 🚀

## Step 1: Get Your API Key

1. Go to <https://lunasentri-web.serverplus.org>
2. Click **Machines** → **Add Machine**
3. Copy your API key

## Step 2: Run This Command

SSH into your Linux server and run:

```bash
curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh | \
  sudo LUNASENTRI_API_KEY="your-api-key-here" \
  LUNASENTRI_SERVER_URL="https://lunasentri-api.serverplus.org" \
  bash
```

**Replace `your-api-key-here` with the API key from Step 1.**

## Step 3: Verify It's Working

```bash
sudo systemctl status lunasentri-agent
```

You should see "active (running)". ✅

## Step 4: Check Your Dashboard

Go to <https://lunasentri-web.serverplus.org/machines>

Your server should appear within 10 seconds! 🎉

---

## Need Help?

**View logs:**

```bash
sudo journalctl -u lunasentri-agent -f
```

**Restart agent:**

```bash
sudo systemctl restart lunasentri-agent
```

**Contact support:**  
<https://github.com/Constantin-E-T/lunasentri/issues>

---

**That's it!** Your monitoring is now live. 🌙✨
