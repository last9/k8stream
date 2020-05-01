package main

const slackTmpl = `
{
	"text": "Last9 Infrastructure Event Alert",
    "blocks": [
        {
            "type": "section",
            "text": {
				"type": "mrkdwn",
				"text": "*Service Alert*\nLast9 K8stream alert.\nPlease refer to <https://github.com/last9/k8stream|K8stream> for configuration options to filter namespace or event-types."
			},
			"accessory": {
        		"type": "image",
        		"image_url": "https://avatars2.githubusercontent.com/u/53378302?s=100&v=4",
        		"alt_text": "Last9 Bot"
      		}
        },
        {{range $i, $msg := .Messages}}
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [
				{
                    "type": "mrkdwn",
                    "text": "*Timestamp*\n{{(formatTime $msg.Timestamp)}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Component*\n{{$msg.Component}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Namespace*\n{{$msg.Namespace}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Event Reason*\n{{$msg.Reason}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Message*\n{{$msg.Message}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Host*\n{{$msg.Host}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Affected Services*\n{{(joinStr $msg.Services "\\n")}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Object Details*\n{{$msg.ReferenceKind}} {{$msg.ReferenceName}}"
                },
				{
					"type": "mrkdwn",
					"text": "*Pod* {{$msg.Pod.name}} | *PodIP* {{$msg.Pod.ip}} | *HostIP* {{$msg.Pod.host_ip}}"
				}
            ]
        },
        {{end}}
		{
      		"type": "context",
      		"elements": [
        		{
          		"type": "mrkdwn",
          		"text": ":eyes: View all details on <https://app.last9.io|Last9>\n:question: Get help at any time with /last9 help"
        		}
      		]
    	}
    ]
}
`
