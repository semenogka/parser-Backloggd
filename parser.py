from selenium import webdriver
from selenium_stealth import stealth

import time

def initDriver():
    driver  = webdriver.Chrome()
    stealth(driver,
            platform="Win32")
    return driver
driver = initDriver() 
driver.get("https://www.ozon.ru")
time.sleep(10)
