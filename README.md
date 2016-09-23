# Netlify Comments

A commenting system for [JAMstack sites](https://jamstack.org).

## How it works

Netlify Comments is both a build tool and a small API.

The API accepts HTTP POST requests to a thread with a JSON body like:

* **POST** /2016/09/this-is-a-thread-on-my-site

```json
{
  "author": "Matt Biilmann",
  "email": "joe@example.com",
  "www": "https://www.example.com",
  "body": "Hi there - this is a fantastic comment!"
}
```

Netlify Comments will check to see that the thread exists and verify that it is
still open, run some checks on the comment to classify obvious spam, and then push
the comment to a Github repository as a JSON document.

That will trigger a build through Netlify with Netlify Comments and a new version
of the thread will be pushed as a JSON file to a static endpoint.

From your site, you can fetch comments and comment metadata from the static endpoint
and let users POST new comments via the API.

Netlify Comments is not a ready made comment system like Disqus or Facebook Comments,
but a buildingblock for making your own custom styled comments on your site.

## Getting Started

### Setting up the static Comments

First clone our [Netlify Comments starter template](https://github.com/netlify/netlify-comments-starter) and push it to your own GitHub account.

Then visit [Netlify](https://app.netlify.com/signup) and pick your new repository. Click **Save** and Netlify will start building your comment threads

### Setting up the API

You'll need to run the API on a server. On the server, we recommend settings these environment variables:

```bash
COMMENT_SITE=https://mysite.exameple.come # URL to your static site
COMMENT_REPO=user/repo # Username/repo of the GitHub repository created from netliy-comments-starter
COMMENT_TOKEN=1253523421313 # A Personal GitHub Access Token with write permissions to the repository
```

With these environment variables in place, run:

```bash
netlify-comment api
```

### Integrating with your site

Each post on your static site that should have comments, needs to add a metadata tag to it's page like this:

```html
<script id="netlify-comments" type="application/json">{"created_at":"2016-07-07T08:20:36Z"}</script>
```

To configure Netlify Comments add a file called `/netlify-comments/settings.json` to your site (this is optional). It should look like this:

```json
{
  "banned_ips": [],
	"banned_keywords": [],
	"banned_emails": [],
	"timelimit": 604800
}
```

These settings controls the rudimentary spam filter and the time limit from a post is created and until
commenting is closed for the thread.

To allow people to comment you'll need your comment form to send a request to the comment API. Here's an
example using the modern `fetch` API:

```js
const thread = document.location.pathname;
fetch(API_URL + thread, {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    author: data.name,
    email: data.email,
    body: data.message,
    parent: data.parent
  })
}).then((response) => {
  console.log("Comment posted!");
});
```

To display comments for a thread, fetch the JSON via:

```js
const slug = document.location.pathname.replace(/\//g, '-').replace(/(^-|-$)/g, '') + '.json';
fetch(COMMENT_URL + '/' + slug).then((response) => {
  console.log("Got comments: %o", response);
});
```

Netlify Comments also builds a file called `threadname.count.json` for each thread with a JSON
object looking like:

```json
{"count": 42}
```

As a lower bandwidth way to fetch comment counts for a thread.

## Licence

Netlify Comments is released under the [MIT License](LICENSE).
Please make sure you understand its [implications and guarantees](https://writing.kemitchell.com/2016/09/21/MIT-License-Line-by-Line.html).
