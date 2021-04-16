#!/usr/bin/env bash

list="event
event.withAnnotations
event_with_opsgenie_priority.check
event_with_opsgenie_priority"

echo "Test open a incident with status 0"
cat event.resolved.json | ./../sensu-opsgenie-handler 


for event in $list ; 
do  
    echo "${event} With default region"
    cat ${event}.json | ./../sensu-opsgenie-handler 
    sleep 5
    cat ${event}.resolved.json | ./../sensu-opsgenie-handler 
done