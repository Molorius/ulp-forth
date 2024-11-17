
#include "esp_log.h"
#include "esp_sleep.h"
#include "ulp.h"
#include "ulp_main.h"

extern const uint8_t ulp_main_bin_start[] asm("_binary_ulp_main_bin_start");
extern const uint8_t ulp_main_bin_end[]   asm("_binary_ulp_main_bin_end");

static void init_ulp(void)
{
    // load the program
    esp_err_t err = ulp_load_binary(0, ulp_main_bin_start,
        (ulp_main_bin_end - ulp_main_bin_start) / sizeof(uint32_t));
    ESP_ERROR_CHECK(err);

    // start the ulp
    err = ulp_run(&ulp_entry - RTC_SLOW_MEM);
    ESP_ERROR_CHECK(err);
}

void app_main(void)
{
    const char *TAG = "app_main";

    ESP_LOGI(TAG, "starting ulp");
    init_ulp();
    ESP_LOGI(TAG, "entering deep sleep");
    esp_deep_sleep_start();
}
