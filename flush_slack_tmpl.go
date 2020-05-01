package main

const slackTmpl = `
{
  "text": "Last9 Infrastructure Event Alert",
  "blocks": [
    {{range $i, $msg := .Messages}}
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*{{(joinStr $msg.Services ",")}}* impacted\nError {{$msg.Message}}\n"
      },
      "accessory": {
        "type": "button",
        "url": "https://app.last9.io",
        "value": "{{$msg.ID}}",
        "style": "primary",
        "text": {
          "type": "plain_text",
          "text": "More Details",
          "emoji": true
        }
      }
    },
    {
      "type": "context",
      "elements": [
        {
          "type": "mrkdwn",
          "text": "*Timestamp* {{(formatTime $msg.Timestamp)}}"
        },
        {
          "type": "mrkdwn",
          "text": "*Component* {{$msg.Component}}"
        },
        {
          "type": "mrkdwn",
          "text": "*Namespace* {{$msg.Namespace}}"
        }
        {{if $msg.ReferenceName}},
        {
          "type": "mrkdwn",
          "text": "*Reference* {{$msg.ReferenceKind}} {{$msg.ReferenceName}}"
        }
        {{end}}
      ]
    },
    {{if $msg.Pod.ip}}
    {
      "type": "context",
      "elements": [
        {
          "type": "mrkdwn",
          "text": "*Pod* {{$msg.Pod.name}}"
        },
        {
          "type": "mrkdwn",
          "text": "*PodIP* {{$msg.Pod.ip}}"
        },
        {
          "type": "mrkdwn",
          "text": "*HostIP* {{$msg.Pod.host_ip}}"
        },
        {
          "type": "mrkdwn",
          "text": "*Host* {{$msg.Host}}"
        }
      ]
    },
    {{end}}
    {
      "type": "divider"
    },
    {{end}}
    {
      "type": "context",
      "elements": [
        {
          "type": "mrkdwn",
          "text": ":eyes: Please refer to <https://github.com/last9/k8stream|K8stream> for configuration options to filter namespace or event-types."
        },
        {
          "type": "mrkdwn",
          "text": ":question: Get help at any time at <https://last9.io|Last9>"
        }
      ]
    }
  ]
}
`
