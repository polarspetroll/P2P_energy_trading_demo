#include "lib_c.h"
#include <Python.h>


///// LCD ////
void LCD_Write(char* str, int i, char* align) {
    Py_Initialize();
    char cmd[300];
    char* start = "from rpi_lcd import LCD\nlcd = LCD()\nlcd.text(\"";
    sprintf(cmd, "%s%s\", %d, \"%s\")", start, str, i, align);
    PyRun_SimpleString(cmd);
    Py_Finalize();
}

void LCD_Clear() {
    Py_Initialize();
    PyRun_SimpleString("from rpi_lcd import LCD\nlcd = LCD()\nlcd.clear()");
    Py_Finalize();
}

void Double_Write(char* str1, char* str2) {
    Py_Initialize();
    char cmd[500];
    char* start = "from rpi_lcd import LCD\nlcd = LCD()\nlcd.text(\"";
    sprintf(cmd, "%s%s\", 1)\nlcd.text(\"%s\", 2)", start, str1, str2);
    PyRun_SimpleString(cmd);
    Py_Finalize();
}