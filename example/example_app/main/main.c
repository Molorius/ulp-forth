/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
#include "driver/uart.h"
#include "driver/gpio.h"
#include "esp_log.h"
#include "ulp.h"
#include "ulp_main.h"

// Used to read the ULP serial out pin,
// wire this to the GPIO the ULP uses.
#define UART_RXD 5
#define UART_TXD UART_PIN_NO_CHANGE
#define UART_RTS UART_PIN_NO_CHANGE
#define UART_CTS UART_PIN_NO_CHANGE
#define UART_PORT 1
#define UART_BAUD 9600 // Use the same baud as the ULP.
#define UART_BUF_SIZE 1024 
#define UART_READ_TASK_SIZE 2048
#define UART_READ_TASK_PRIORITY 10

static void uart_read_task(void *arg)
{
    const char *TAG = "uart_read_task";
    uint8_t *data = (uint8_t *) malloc(UART_BUF_SIZE);

    ESP_LOGI(TAG, "starting");
    for (;;) {
        // read from the uart
        int len = uart_read_bytes(UART_PORT, data, (UART_BUF_SIZE - 1), 20 / portTICK_PERIOD_MS);
        if (len) { // if we received anything
            data[len] = 0; // put a null byte at the end
            int good_start = 0;
            int good_end = 0;
            // print all substrings that end with a null character
            for (int i=0; i<len+1; i++) {
                if (data[i] == 0) {
                    good_end = i;
                    int good_len = good_end - good_start;
                    if (good_len > 0) {
                        printf("%s", &data[good_start]);
                    }
                    good_start = i+1;
                }
            }
        }
    }
}

static void init_uart(void)
{
    uart_config_t uart_config = {
        .baud_rate = UART_BAUD,
        .data_bits = UART_DATA_8_BITS,
        .parity    = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE,
        .source_clk = UART_SCLK_DEFAULT,
    };
    ESP_ERROR_CHECK(uart_driver_install(UART_PORT, UART_BUF_SIZE * 2, 0, 0, NULL, 0));
    ESP_ERROR_CHECK(uart_param_config(UART_PORT, &uart_config));
    ESP_ERROR_CHECK(uart_set_pin(UART_PORT, UART_TXD, UART_RXD, UART_RTS, UART_CTS));

    xTaskCreate(uart_read_task, "uart_read_task", 2048, NULL, UART_READ_TASK_PRIORITY, NULL);
}

extern const uint8_t ulp_main_bin_start[] asm("_binary_ulp_main_bin_start");
extern const uint8_t ulp_main_bin_end[]   asm("_binary_ulp_main_bin_end");

static void init_ulp(void)
{
    const char *TAG = "init_ulp";
    // load the program
    esp_err_t err = ulp_load_binary(0, ulp_main_bin_start,
        (ulp_main_bin_end - ulp_main_bin_start) / sizeof(uint32_t));
    ESP_ERROR_CHECK(err);

    // start the ulp
    ESP_LOGI(TAG, "starting ulp");
    err = ulp_run(&ulp_entry - RTC_SLOW_MEM);
    ESP_ERROR_CHECK(err);
}

void app_main(void)
{
    const char *TAG = "app_main";

    ESP_LOGI(TAG, "starting ulp example app");

    init_uart();
    init_ulp();
}
