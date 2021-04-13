Amaze crash automation
---
Use this API function as a gateway to create any GitHub issue.

The flow is as follows:
- User reports a crash either through email or through telegram bot
- These crash requests are handled by amaze-telegram-bot or amaze-crashreports-webhook respectively.
- In order to check for existing issue reported, both these bots call this API function
- Here we're making various checks, like checking version support or any unofficial crash report.
- This API function then calls page wise GitHub issues of AmazeFileManager repo and checks if crash report matches the reported one
- If it can't find any report, in-turn calls amaze-github-automation to create a new issue
- Returns a response for email reply or telegram bot reply.

Request
---
POST request - localhost:10000
Query params
- channel : (given channel, email or telegram)
- token : (api function token for authentication, matches with the one set in env variables)

Body:
```
{
    "title": "test title",
    "body": "test body"
}
```

Response:
```
{
    "number": 2409,
    "html_url": "https://github.com/TeamAmaze/AmazeFileManager/issues/2409",
    "body": "crash",
    "isUnofficial": false
}
```

Note:
---
Before running this API function make sure to set the environment variables in your pipelines and remove from this script.