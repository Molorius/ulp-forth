/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
#include "driver/uart.h"
#include "driver/gpio.h"
#include "esp_private/rtc_ctrl.h"
#include "soc/rtc_cntl_reg.h"
#include "esp_log.h"
#include "ulp.h"
#include "ulp_main.h"

#define ULP_WAKEUP_PERIOD_US 1000000

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
#define WAKE_TASK_SIZE 2048
#define WAKE_TASK_PRIORITY 9

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

    xTaskCreate(uart_read_task, "uart_read_task", UART_READ_TASK_SIZE, NULL, UART_READ_TASK_PRIORITY, NULL);
}

SemaphoreHandle_t ulp_wake_sem = NULL;

static void ulp_wake_isr(void *arg)
{
    BaseType_t yield = 0;
    xSemaphoreGiveFromISR(ulp_wake_sem, &yield);
    if (yield) {
        portYIELD_FROM_ISR();
    }
}

static void ulp_wake_task(void *arg)
{
    const char *TAG = "ulp_wake_task";

    ESP_LOGI(TAG, "waiting for the ulp to wake us");
    for (;;) {
        if (xSemaphoreTake(ulp_wake_sem, portMAX_DELAY) == pdTRUE) {
            ESP_LOGI(TAG, "ulp used the wake instruction");
        }
    }
}

static void init_wake(void)
{
    ulp_wake_sem = xSemaphoreCreateBinary();
    if (ulp_wake_sem == NULL) {
        ESP_ERROR_CHECK(ESP_ERR_NO_MEM);
    }
    
    // set up the isr
    ESP_ERROR_CHECK(rtc_isr_register(&ulp_wake_isr, NULL, RTC_CNTL_SAR_INT_ST_M, 0));
    // allow the interrupt to occur
    REG_SET_BIT(RTC_CNTL_INT_ENA_REG, RTC_CNTL_ULP_CP_INT_ENA_M);
    xTaskCreate(ulp_wake_task, "ulp_wake_task", WAKE_TASK_SIZE, NULL, WAKE_TASK_PRIORITY, NULL);
    
}

extern const uint8_t ulp_main_bin_start[] asm("_binary_ulp_main_bin_start");
extern const uint8_t ulp_main_bin_end[]   asm("_binary_ulp_main_bin_end");

static void init_ulp(void)
{
    const char *TAG = "init_ulp";
    // load the program
    ESP_LOGI(TAG, "loading program");
    esp_err_t err = ulp_load_binary(0, ulp_main_bin_start,
        (ulp_main_bin_end - ulp_main_bin_start) / sizeof(uint32_t));
    ESP_ERROR_CHECK(err);

    // set wakeup period
    ESP_LOGI(TAG, "setting ulp wakeup period to %d microseconds", ULP_WAKEUP_PERIOD_US);
    ESP_ERROR_CHECK(ulp_set_wakeup_period(0, ULP_WAKEUP_PERIOD_US));

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
    init_wake();
    init_ulp();
}
