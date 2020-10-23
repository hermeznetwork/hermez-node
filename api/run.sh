#!/bin/sh -e

USAGE="Available options:
    doc     Start documentation UI at http://localhost:8001 and the mock up server at http://localhost:4010
    mock    Start the mock up server at http://localhost:4010
    editor  Start the documentation editor at  http://localhost:8002
    stop    Stop all runing services started using this script
    help    display this message"

case "$1" in
    doc)
        sudo docker-compose up -d hermez-api-doc hermez-api-mock && echo "\n\nStarted documentation UI at http://localhost:8001 and mockup server at http://localhost:4010"
        ;;
    mock)
        sudo docker-compose up -d hermez-api-mock && echo "\n\nStarted mockup server at http://localhost:4010"
        ;;
    editor)
        sudo docker-compose up -d hermez-api-editor hermez-api-mock && echo "\n\nStarted spec editor at http://localhost:8002 and mockup server at http://localhost:4010"
        ;;
    stop)
        sudo docker-compose rm -sf && echo "\n\nStopped all the services initialized by this script"
        ;;
    help)
        echo "$USAGE"
        ;;
    *)
        echo "Invalid option.\n\n$USAGE"
        ;;
    esac