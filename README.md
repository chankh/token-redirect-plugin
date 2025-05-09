# Google Cloud Media CDN MPEG-DASH Cookieless Token Authentication Plugin

This repository contains a Google Cloud Media CDN plugin that enables cookieless token authentication for MPEG-DASH assets. While Media CDN natively supports cookieless authentication for HLS content, this plugin extends that functionality to MPEG-DASH by leveraging HTTP redirection to rewrite tokens from query parameters into URL paths.

## Overview

This plugin addresses the limitation of Media CDN's native cookieless support, which is currently restricted to HLS. For MPEG-DASH content, this plugin implements a workaround by:

1.  **Receiving requests with tokens in query parameters:** The client initially requests the MPEG-DASH manifest (e.g., `manifest.mpd?token=...`).
2.  **Rewriting tokens into the URL path:** The plugin intercepts the request and redirects it, moving the token from the query parameters to a specific path segment (e.g., `/token/.../manifest.mpd`).
3.  **Media CDN handles the rewritten URL:** Media CDN can then process the request, extracting the token from the path and validating it using a standard token authentication plugin.

This approach allows for cookieless authentication of MPEG-DASH content within the Media CDN ecosystem, providing enhanced security and compatibility.

## Features

*   **Cookieless Authentication for MPEG-DASH:** Enables token-based authentication without relying on cookies.
*   **HTTP Redirection for Token Rewriting:** Uses HTTP redirects to move tokens from query parameters to URL paths.
*   **Seamless Integration with Media CDN:** Designed to work as a plugin within the Media CDN environment.
*   **Improved Security:** Eliminates the security risks associated with cookie-based authentication.
*   **Enhanced Compatibility:** Works across various client environments without cookie limitations.

## Prerequisites

*   **Google Cloud Account:** You need a Google Cloud account with Media CDN enabled.
*   **Media CDN Configuration:** You should have a Media CDN setup with an origin and a mapping.
*   **Basic Understanding of Media CDN Plugins:** Familiarity with how to deploy and configure plugins in Media CDN.
*   **Go Programming Language Knowledge:** This sample code is written in a Go. You'll need knowledge of Go to understand and modify the code.

## Getting Started

1.  **Clone the Repository:**
    ```bash
    git clone <repository-url>
    cd <repository-directory>
    ```

2.  **Build and Upload the Plugin**
    *   Modify the variables defined in the first few lines in [Makefile](Makefile) to your environment specific.
    *   Run the command `make gar` to create a repo in Google Cloud Artifact Registory for the plugin.
    *   Run the command `make all` to build the docker image.
    *   Finally run `make docker-push` to push the docker image to Google Cloud Artifact Registry.

3.  **Create Media CDN Service Extension**
    *   Enable the APIs.
    ```bash
    gcloud services enable networkservices.googleapis.com
    gcloud services enable networkactions.googleapis.com
    ```
    *   Run `make wasm` to create the Service Extension Plugin and Wasm Action for use with Media CDN.

4.  **Attach the Plugin to Media CDN:**
    *   Follow the Google Cloud Media CDN documentation to Create your Media CDN [origin](https://cloud.google.com/media-cdn/docs/quickstart#create-origin) and [service](https://cloud.google.com/media-cdn/docs/quickstart#create-service).
    *   Configure your Media CDN routes to use this plugin. Update the routes in the configuration to add the `wasmAction` header as shown in the [official guide](https://cloud.google.com/service-extensions/docs/attach-plugins-to-routes#attach-plugin).

5.  **Test Your Implementation:**
    *   Use tools like `curl` or a browser to test content access with and without the correct tokens in the query parameters.
    *   Verify that the plugin correctly redirects requests and that the subsequent token validation plugin allows or blocks access as expected.

## Code Structure

*   `main.go`: The main plugin file containing the redirection logic.
*   `main_test.go`: Unit tests.
*   `README.md`: This documentation file.

