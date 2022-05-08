# Generate Aysnc API Doco

https://www.asyncapi.com/tools/generator

    docker pull asyncapi/generator
    docker run --rm -it -v ${PWD}/docs:/app \
        asyncapi/generator \
        pw_ws_broker.async-api.yaml \
        @asyncapi/html-template \
        -o pw_ws_broker


-or- if you want it installed locally

    npm install -g @asyncapi/generator

and then
    
    ag docs/pw_ws_broker.async-api.yaml \
        @asyncapi/html-template \
        -o docs/pw_ws_broker/