# Golang Request Demo Page

A webserver to forward form contact requests to your email address.

## Installation

### Installation using Docker:

1. Clone the repo: `git clone ...`
2. Move the env.list.example file to env.list: `cp env.list.example env.list`
3. Replace the variables in `env.list` with your credentials and the email addresses you want the email to be forwarded to.
4. Build the Docker image: `docker build -t golang-request-demo-page .`
5. Deploy the image: `docker run --env-file env.list --publish 8080:8080 --detach --name golang-request-demo-page --rm golang-request-demo-page`
    - This will run the image in the background. Remove `--detach` if you want to see the output log

To stop the container: `docker stop golang-request-demo-page`
