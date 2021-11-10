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
	"relays": [] // Relay Pins. Eg : [40, 38, 10]
}
```
