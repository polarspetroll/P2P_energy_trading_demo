##### Requirements 

- Python C API
- Python rpi_lcd library
- Python ina_219 library
- go github.com/polarspetroll/gopio package


##### Hardware Requirements

- INA219 Module
- Raspberrypi (3 or 4)
- Relay
- 16x2 Character LCD
- Solar Panel
- Li-Po Battery Charger Module


---

- ina 
	- gcc test.c lib_c.c -I /usr/include/python3.7/ -lpython3.7 $(python3.7-config --includes --ldflags)