## sync-service

### send request
1. Create a json payload with the following structure,
2. Create HMAC signed JWT using the "SYNC_SECRET" env variable you've specified in the server configuration.
3. Send it as the body in a POST request to /sync on the server instance.

```json
{
  "src": {
		"server":   "SOURCESERVER"
		"username": "sourceuser"
		"password": "sourcepassword"
		"domain":   "YOURDOMAIN"
		"share":    "sourceshare$"
		"path":     "path/on/share"
	}
	"src": {
		"server":   "DESTINATIONSERVER"
		"username": "destinationuser"
		"password": "destinationpassword"
		"domain":   "YOURDOMAIN"
		"share":    "destinationshare$"
		"path":     "path/on/share"
	}
}
```
