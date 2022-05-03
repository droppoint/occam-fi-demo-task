# occam.fi demo task

App that connects to several sources of financial data and merges streams. Currently supported APIs:
- EODHistoricalData
- Exmo.com

### Usage
    go build .
    ./occam-fi-demo-task BTCUSD

### Solution
My solution consists of three abstractions - _TriggerArregator_, _Subscriber_ and _Client_. _Client_ implements
methods required to connect to API. _Subscriber_ via _Client_ connects to API and spawns goroutine that reads messages
and sends them through the channel. _TriggerArregator_ spawns a goroutine for each _Subscriber_ which reads messages
from channel and saves last price in a map. _TriggerArregator_ also spawns one goroutine that reads last prices from
map, calculates the average and outputs it to the console. I decided to slightly change interface for
SubscribePriceStream and add a context.Context argument, it was necessary for graceful shutdown of goroutines in
_Subscriber_. 
