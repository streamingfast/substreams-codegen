# Developing a codegen conversation

## Usage

Use the codegen.substreams.dev and codegen-staging.substreams.dev endpoints.

Invoke with:

```bash
substreams init --discovery-endpoint http://localhost:9000
```

## Develop

Run the `substreams-codegen` backend:

```bash
# remove the DEBUG var if you want info level logs
DEBUG=.* go run ./cmd/substreams-codegen api --http-listen-addr "*:9000"
```


## Principles

You write a `Conversation` (or `Convo` for short) struct.

- This function embed or has a `State` variable with the state you want to build.
  - This _State_ is what gets serialized to JSON and sent to the client. It must be a usable interface, so pick your names wisely.
  - Anything that is exported will be shared with the client, and come back during hydration (if a connection gets interrupted, or the user kept their `generator.json` state, and wants to simply rebuild).
- It has an `Update()` method that is registered into the `loop.EventLoop` by the `codegen` framework.
  - This function can mutate state, and it is the only method who can mutate state.
  - This function follows the ELM paradigm, and must never do long-running operations, only quick routing.
  - A `NextStep()` function will choose where to go based on the current state, and continue the conversation to make sure it is valid and completely filled in.
  - Any long running work is done in a `loop.Cmd` (that runs async), and that Cmd returns a message for error handling or continuation.
  - Any loops are done by sending a `loop.Cmd` that does an iteration, and returns a message to continue the loop. The ending condition is merely the rescheduling of that same Cmd (or not, to end the loop).

The code generation:

- Any `gotmpl` files will go through templating, and be passed the _State_ struct as a single parameter.
- The _State_ struct should have helper methods to allow getting data from the state

## Some notes on popular contracts

0x1f98431c8ad98523631ae4a59f267346ea31f984 -> Uniswap V3 Factory
https://api.etherscan.io/api?module=contract&action=getabi&address=0xe592427a0aece92de3edee1f18e0157c05861564 -> Uniswap V3 Router
https://info.uniswap.org/#/tokens/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48 -> USDC Contract -> https://etherscan.io/token/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
