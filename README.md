# How to setup this app

## Pre-requisites
Everything listed is free:  
  - A github account.  
  - Your uattend credentials.
  - A new Discord server and a webhook for it (To notify you of successful runs or errors)  
  - Your Discord user ID

## Creating a discord server's webhook
1. Select the newly created server.
2. Click on the server dropdown on the top left next to the discord icon.
3. Select `Server Settings`. (3rd option)
4. Select `Integrations` on the left menu, then `webhooks`.
5. Click on `New Webhook`, give this "bot" a different name if you want to.
6. Select the text channel for which it will post messages at.
7. Click `Copy Webhook URL` and paste it somewhere. You will need this later.


## Getting your Discord user ID
In your server's text channel, try to `@` yourself but then add a `\` before the `@` and submit the message.  
The message should look like this: `<@!12345678904564>`  
You will need this later.  


Main steps:
1. Head to https://github.com/JohnnyLin-a/uattend-automator while logged in.
2. Click on the top right button `Fork`.
3. Once the fork is complete, Click on the `Actions` tab.
4. Click on the `I understand my workflows, go ahead nad enable them` button.
5. Click on `run-app` under Workflows and click on `Enable workflow`.  
6. Now click on the `Settings` tab. (same row as `Actions`)
7. Click on `Secrets` and click on `New repository secret`.
8. For the secret's name type in `UATTEND_CONFIG`.
9. For the secret's value, fill in the following template and copy paste it into the text box and then `Add secret`.  
    Replace these values:
      - `YOUR_USERNAME`: Your username
      - `YOUR_PASSWORD`: Your password
      - `YOUR_ORGANIZATION_NAME`: Your organization name
      - `YOUR_WEBHOOK`: Your Discord webhook URL
      - `YOUR_DISCORD_USER_ID`: Your Discord User ID

Template:
```
{
    "Credentials": {
        "Login": "YOUR_USERNAME",
        "Password": "YOUR_PASSWORD"
    },
    "OrgURL": "https://v2.trackmytime.com/YOUR_ORGANIZATION_NAME",
    "Workdays": [1,2,3,4,5],
    "behavior": {
        "PunchType": "Benefit",
        "InTime": "",
        "OutTime": "",
        "BenefitType": "OTH - Other",
        "BenefitHours": "8.00"
    },
    "Discord": {
        "Webhook": "YOUR_WEBHOOK",
        "Mention": "YOUR_DISCORD_USER_ID"
    }
}
```


Your automator is now fully set-up.  
The automator runs every week on Mondays at 13:00 UTC+00.  
Once the automator is done, it will send a discord message to that text channel to notify how many rows it has automated, and/or if there was any errors.