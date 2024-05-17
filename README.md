# SlothChain 🦥

Forked from Wasmd and changed to Rollkit (https://rollkit.dev/tutorials/cosmwasm).

For a production version of this we will obviously add Wasmd as a normal dependency rather than forking it.

## How is SlothChain Lazy?

SlothChain is meme chain. The goal is to make it very memey. It is a chain for the lazy ones.

It is implemented as a Rollkit chain, with CosmWasm and some contracts such as DAODAO to make it useful.

![SlothChain](slothchain.png)

### Sloth NFTs
The Celestine Sloth Society NFTs (just called "Sloths" from here on) are currently on Stargaze.
With IBC and ICS721 they can be transferred to any chain. Such as SlothChain.

The Sloths are lazy, they don't do much. They just hang around and occasionally do something. They are not very useful, but they are cute.

Also, they can be used for governance:

### Proof of Laziness
Governance is controlled by the lazy ones. The chain has DAODAO where you can stake your Sloth to participate in governance.

Proof of Laziness should also get you LazyCoin or something, but this is not fully implemented yet.

(As a side not: we could implement a simple contract that swaps Sloths for LazyCoin at a constant rate, but this has not been implemented yet)

### The chain itself is lazy
* The logger has sloths in the formatting: `🦥 2:37PM INF starting node with Rollkit in-process`
* The chain is running with the Rollkit lazy aggregator, so it only creates blocks when there are transactions
* The base token is `ulazy`

### Future plans
If we get some funding to actually see this project through I would like to:
* Add a bunch of stupid memey things to the chain like sloth emojies everywhere, slower blocks, etc (just really lean into the lazy theme)
* Make this a fun governance project where the Sloth community can come together and own something (lazily, of course)
* Put some serious degen stuff in here, but make it lazy somehow. Maybe some crosschain stuff can make this happen.
* Just more fun stuff, really. Fun, lazy and memey.