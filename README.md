##### Hardware Requirements

- INA219 Module
- Raspberrypi (3 or 4)
- Relay
- 16x2 Character LCD
- Solar Panel
- Li-Po Battery Charger Module


---

###### Installation

```
git clone https://github.com/polarspetroll/P2P_energy_trading_demo.git
make install
make build
make run
```

###### Configuration

- config.json
```json
{
	"relays": [40], //relay pins
	"interval": "1s", // interval sleep for consumption 
	"ina_addresses": ["0x40"] // available addresses for INA219 module
}
```
