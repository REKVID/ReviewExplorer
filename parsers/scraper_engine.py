import re
import time
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys


def parse_review(text):
    data = {"date": None, "author": None, "rating": None, "text": text}
    
    m = re.search(r'(\d{4}-\d{2}-\d{2})', text)
    if m: data["date"] = m.group(1)

    if "Положительный" in text: data["rating"] = "positive"
    elif "Отрицательный" in text: data["rating"] = "negative"
    else: data["rating"] = "neutral"

    m = re.search(r'Это ложь\s+\d+\s*\n(.*?)(Ответить|$)', text, re.S)
    if m: data["text"] = m.group(1).strip()
    
    return data

def scrape_reviews(school_name):
    options = Options()
    options.add_argument("--headless")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--disable-gpu")
    options.add_argument("user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    
    options.binary_location = "/usr/bin/chromium"
    service = Service(executable_path="/usr/bin/chromedriver")
    
    driver = webdriver.Chrome(service=service, options=options)
    
    reviews = []
    try:
        driver.get("https://ya.ru")
        time.sleep(3)

        search = driver.find_element(By.NAME, "text")
        search.send_keys(f"schoolotzyv {school_name}")
        search.send_keys(Keys.RETURN)
        time.sleep(4)

        link = None
        for a in driver.find_elements(By.CSS_SELECTOR, "a[href]"):
            href = a.get_attribute("href")
            if href and "schoolotzyv.ru" in href:
                link = href
                break
        
        if not link:
            return []

        driver.get(link)
        time.sleep(5)
        
        driver.execute_script("window.scrollTo(0, document.body.scrollHeight/2);")
        time.sleep(2)
        
        blocks = driver.find_elements(By.CSS_SELECTOR, "div.comments-content-container")
        
        for b in blocks:
            text = b.text
            parsed = parse_review(text)
            if not parsed.get("text") or len(parsed["text"]) < 5:
                parsed["text"] = text[:500] 
            reviews.append(parsed)

    except Exception as e:
        pass
    finally:
        driver.quit()
        
    return reviews
