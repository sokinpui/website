---
title: "how-to-setup-obsidian-livesync"
desc: ""
createdAt: "2025-12-23T11:34:21Z"
---

Here is a summary guide to setting up the **Obsidian Self-hosted LiveSync** plugin on a Synology NAS.

---

# **Setup Guide for Synology NAS**

**Prerequisite:** Install **Container Manager** (formerly Docker) from the Synology Package Center.

## **Step 1: File System Setup**

1.  Open **File Station** on your Synology.
2.  Navigate to your `docker` share.
3.  Create a new folder named `obsidian-livesync`.

## **Step 2: Create the Container (Docker Compose)**

1.  Open **Container Manager**.
2.  Go to the **Project** tab and click **Create**.
3.  **Project Name:** `obsidian-livesync`
4.  **Path:** Select the `docker/obsidian-livesync` folder you created.
5.  **Source:** Select "Create docker-compose.yml".
6.  Paste the following configuration into the editor:

```yaml
version: "3.9"
services:
  couchdb:
    image: couchdb:3.3.3
    container_name: obsidian-livesync
    restart: always
    environment:
      - COUCHDB_USER=admin_user # CHANGE THIS
      - COUCHDB_PASSWORD=strong_password # CHANGE THIS
    volumes:
      - ./data:/opt/couchdb/data
      - ./local.d:/opt/couchdb/etc/local.d
    ports:
      - 5984:5984
```

_Tip: Change `admin_user` and `strong_password` to your preferred credentials._

7.  Click **Next** and finish the wizard to start the container.

## **Step 3: Initialize the Database**

Because CouchDB starts empty, you must create the database and configure it. The easiest way is via the web interface (Fauxton).

1.  Open a web browser and go to `http://<YOUR_NAS_IP>:5984/_utils`.
2.  Login with the username/password you set in the Docker Compose file.
3.  Click **"Create Database"** in the top right.
4.  Name it `obsidian` (or `obsidiandb`).
5.  Select **Non-partitioned** and click Create.

_Optional but Recommended:_ You may need to enable CORS if you have trouble connecting from devices. In the CouchDB UI, go to **Configuration** -> **CORS** and enable it for all domains (`*`) or your specific device IPs.

## **Step 4: Configure Obsidian Plugin**

1.  In Obsidian, install the **Self-hosted LiveSync** plugin (Community Plugins).
2.  Go to **Settings > Self-hosted LiveSync**.
3.  **Remote Type:** Select `CouchDB`.
4.  **URI:** `http://<YOUR_NAS_IP>:5984`
5.  **Username / Password:** Enter the credentials from your Docker Compose file.
6.  **Database Name:** `obsidian` (or whatever you named it in Step 3).
7.  Click **Test** (should say "Connected").
8.  Click **Check** (fixes database configuration automatically).
9.  Click **Apply Settings**.
10. use `live-sync` sync mode

## **Step 5: Syncing Other Devices**

Once the first device is set up:

1.  Go to the **"Setup"** tab in the plugin settings on the working device.
2.  Click **"Copy setup URI"**.
3.  On your _new_ device (e.g., phone), install the plugin.
4.  Open the plugin settings and look for **"Connect with Setup URI"**.
5.  Paste the code. This copies all settings instantly.
6.  use `live-sync` sync mode
