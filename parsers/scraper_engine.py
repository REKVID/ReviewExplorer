import re
import time
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
import shutil

def parse_review(text):
    data = {"date": None, "rating": "neutral", "text": text}
    
    # Дата
    m = re.search(r'(\d{4}-\d{2}-\d{2})', text)
    if m: data["date"] = m.group(1)

    # Рейтинг
    if "Положительный" in text: data["rating"] = "positive"
    elif "Отрицательный" in text: data["rating"] = "negative"
    else: data["rating"] = "neutral"

    # Извлекаем основной текст
    m = re.search(r'Это ложь\s+\d+\s*\n(.*?)(Ответить|$)', text, re.S)
    if m: data["text"] = m.group(1).strip()
    
    return data

def scrape_reviews(school_name):
    options = Options()
    options.add_argument("--headless")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--disable-gpu")
    options.add_argument("--disable-blink-features=AutomationControlled")
    
    chrome_path = shutil.which("chromium") or shutil.which("google-chrome") or "/usr/bin/chromium"
    options.binary_location = chrome_path
    
    try:
        service = Service(executable_path=shutil.which("chromedriver") or "/usr/bin/chromedriver")
        driver = webdriver.Chrome(service=service, options=options)
    except:
        driver = webdriver.Chrome(options=options)
    
    reviews = []
    try:
        driver.get("https://ya.ru")
        time.sleep(2)

        try:
            search = driver.find_element(By.NAME, "text")
        except:
            search = driver.find_element(By.CSS_SELECTOR, "input")
            
        search.send_keys(f"schoolotzyv {school_name}")
        search.send_keys(Keys.RETURN)
        time.sleep(5)

        # Ищем ссылку именно на страницу школы
        link = None
        for a in driver.find_elements(By.CSS_SELECTOR, "a[href*='schoolotzyv.ru']"):
            href = a.get_attribute("href")
            if href and "schoolotzyv.ru/schools" in href:
                link = href
                break
        
        if not link:
            return []

        driver.get(link)
        time.sleep(5)
        
        # Скролл для подгрузки (как в save.py)
        driver.execute_script("window.scrollTo(0, document.body.scrollHeight/2);")
        time.sleep(2)
        
        # Ищем элементы отзывов по селекторам из save.py
        blocks = driver.find_elements(By.CSS_SELECTOR, ".comment-item, .review-item, [class*='comment']")
        
        for b in blocks:
            text = b.text.strip()
            if len(text) < 50: continue
            
            parsed = parse_review(text)
            if parsed["text"]:
                reviews.append(parsed)

    except Exception as e:
        print(f"Error scraping: {e}")
    finally:
        driver.quit()
        
    return reviews
