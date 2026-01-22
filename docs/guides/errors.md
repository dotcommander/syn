# Error Handling Guide

This guide covers common errors you may encounter when using `syn`
and how to resolve them.

## Common Errors

### API key is not configured

**Error Message:**

```text
API key is not configured. Set SYN_API_KEY or configure in ~/.config/syn/config.yaml
```

**Cause:**

The `SYN_API_KEY` environment variable is not set or no API key is
configured in the config file.

**Resolution:**

1. Set the `SYN_API_KEY` environment variable:

   ```bash
   export SYN_API_KEY=your_api_key_here
   ```

2. Or configure it in `~/.config/syn/config.yaml`:

   ```yaml
   api_key: your_api_key_here
   ```

---

### failed to read file

**Error Message:**

```text
failed to read file /path/to/file: open /path/to/file: no such file or directory
```

**Cause:**

The specified file path does not exist or is not readable.

**Resolution:**

1. Verify the file path is correct
2. Check file permissions: `ls -la /path/to/file`
3. Use absolute paths if relative paths are not resolving correctly

---

### API error: 401

**Error Message:**

```text
API error: 401 - {"error":"Unauthorized"}
```

**Cause:**

The API key is invalid, expired, or malformed.

**Resolution:**

1. Verify your API key is correct
2. Check if the key has expired
3. Ensure there are no extra spaces or characters in the key
4. Regenerate the API key if necessary

---

### API error: 429

**Error Message:**

```text
API error: 429 - {"error":"Too Many Requests"}
```

**Cause:**

Rate limit exceeded. The API has received too many requests in a short
period.

**Resolution:**

1. Wait a few seconds before retrying
2. The client will automatically retry with exponential backoff
3. Configure retry settings in `~/.config/syn/config.yaml`:

   ```yaml
   retry_config:
     max_attempts: 5
     initial_backoff: 2s
     max_backoff: 60s
   ```

---

### failed to send request

**Error Message:**

```text
failed to send request: Post "https://api.example.com/chat/completions":
dial tcp: lookup api.example.com: no such host
```

**Cause:**

DNS resolution failure or network connectivity issue.

**Resolution:**

1. Check your internet connection
2. Verify the API base URL in your config is correct
3. Try pinging the host: `ping api.example.com`
4. Check if a VPN or proxy is interfering

---

### API error: 500 / 502 / 503 / 504

**Error Message:**

```text
API error: 503 - Service Unavailable
```

**Cause:**

The API server is experiencing issues or is temporarily unavailable.

**Resolution:**

1. The client will automatically retry with exponential backoff
2. Wait a few minutes and try again
3. Check the API status page if available
4. If the issue persists, contact API support

---

### no choices in response

**Error Message:**

```text
no choices in response
```

**Cause:**

The API returned a response without any completion choices, possibly due
to an invalid model or request format.

**Resolution:**

1. Verify the model name is correct: `syn model list`
2. Check if the model is available in your region
3. Try with a different model using the `--model` flag

---

### no texts provided for embedding

**Error Message:**

```text
no texts provided for embedding
```

**Cause:**

The embed command was called without providing any text input.

**Resolution:**

1. Provide text input via stdin or file:

   ```bash
   echo "your text here" | syn embed
   syn embed < input.txt
   ```

---

### failed to read image file

**Error Message:**

```text
failed to read image file: open image.png: no such file or directory
```

**Cause:**

The image file specified for vision analysis does not exist.

**Resolution:**

1. Verify the image file path is correct
2. Check the file exists: `ls -la image.png`
3. Use absolute paths if needed
4. Ensure the image format is supported (PNG, JPEG, GIF, WebP)

---

## Debug Mode

Enable verbose logging to get more detailed error information:

```bash
export SYN_VERBOSE=1
syn chat "your prompt"
```

Or in `~/.config/syn/config.yaml`:

```yaml
verbose: true
```

## Getting Help

If you encounter an error not covered in this guide:

1. Enable verbose mode to see detailed logs
2. Check the API documentation for the specific endpoint
3. Report issues at: <https://github.com/vampire/syn/issues>
