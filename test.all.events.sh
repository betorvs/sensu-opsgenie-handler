#!/usr/bin/env bash

list="event
event.withAnnotations
event_with_opsgenie_priority.check
event_with_opsgenie_priority"

for event in $list ; 
do  
    echo "${event} With URL https://api.opsgenie.com/"
    cat ${event}.json | ./sensu-opsgenie-handler --url 'https://api.opsgenie.com/'
    sleep 5
    cat ${event}.resolved.json | ./sensu-opsgenie-handler --url 'https://api.opsgenie.com/'
done
    sleep 5
for event in $list ; 
do  
    echo "${event} With URL https://api.opsgenie.com"
    cat ${event}.json | ./sensu-opsgenie-handler --url 'https://api.opsgenie.com'
    sleep 5
    cat ${event}.resolved.json | ./sensu-opsgenie-handler --url 'https://api.opsgenie.com'
done
    sleep 5
for event in $list ; 
do  
    echo "${event} Without URL"
    cat ${event}.json | ./sensu-opsgenie-handler
    sleep 5
    cat ${event}.resolved.json | ./sensu-opsgenie-handler
done