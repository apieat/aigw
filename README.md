# Aigw
Apieat AI GW

AI Gateway to expose your api to openapi by its `function call` function. 
Work as a Microservice, so anyone can deploy it with his site.

# function target

- [x] provide a completion api by wrapping openai completion api
- [x] transfer completion to openai with function list generated by openapi doc
- [x] parse openai function call to api calling
- [ ] transfer api result to openai and get the result
- [ ] response the result to user

# Usage

## config

```
openai:
  token: your token
  ask_ai_to_analyse_result: true/false
  sync: true/false
api:
  def: ./def.yaml
  base_url: your server's base url
```

sync: make calling running in sync mode or not, sync mode will response result of ai
ask_ai_to_analyse_result: if you want to ask ai to analyse api calling result, set this argument to true, this will only work in sync mode.

Run this micro-service need a openai token and a openapi definition.

AIGW will transform all your openapi operation to function list which will send will prompt to openai. so when openai response with function_call, AIGW will call your operation as http request.

## api

post /completion

|  content-type    |  | application/json (required) |
| ---- | ------------ | ---------------- |
|   request body   |  | object           |
|      | prompt | the prompt which will send to openai to do completion |
| | id | request id which will send back to original server when openai response |

# Make a openapi definition

Try our editor online : [Apieat](https://apieat.com)

中文用户：[百家饭OpenAPI平台](https://rongapi.cn)
