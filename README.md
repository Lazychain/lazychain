# SlothChain 🦥

Forked from Wasmd and changed to Rollkit (https://rollkit.dev/tutorials/cosmwasm).

For a production version of this we will obviously add Wasmd as a normal dependency rather than forking it.

## How is SlothChain Lazy?

### Sloth NFTs
TODO: Write about this

### The chain itself is lazy
* The logger has sloths in the formatting: `🦥 2:37PM INF starting node with Rollkit in-process`
* The chain is running with the Rollkit lazy aggregator, so it only creates blocks when there are transactions