# airdrip

## About
Airdrop but with more drip. Please excuse the name, the one I originally wanted is already claimed. Maybe I'll get another idea at some point.

## Disclaimer
This project is still in progress. It may take a while for me to finish it.

## Want to test the signaling service?

**`In a first Terminal`**
```
git clone git@github.com:JakWai01/airdrip.git`
make server
```

**`In another Terminal`**
```
export MAC="123"
export COMMUNITY="a"
make client
```

**`In a third Terminal`**
```
export MAC="124"
export COMMUNITY="a"
make client
```