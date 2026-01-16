import re, time, shutil
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys

def scrape_reviews(school_name):
    opts = Options()
    for arg in ["--headless", "--no-sandbox", "--disable-dev-shm-usage", "--disable-gpu", "--disable-blink-features=AutomationControlled"]:
        opts.add_argument(arg)
    opts.add_experimental_option("excludeSwitches", ["enable-automation"])
    opts.add_experimental_option('useAutomationExtension', False)
    opts.add_argument("user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    chrome_path = shutil.which("chromium") or shutil.which("google-chrome") or "/usr/bin/chromium"
    if chrome_path: opts.binary_location = chrome_path
    
    try:
        driver = webdriver.Chrome(service=Service(shutil.which("chromedriver") or "/usr/bin/chromedriver"), options=opts)
        driver.execute_script("Object.defineProperty(navigator, 'webdriver', {get: () => undefined})")
    except:
        return []
    
    reviews = []
    try:
        driver.get("https://ya.ru")
        time.sleep(3)
        
        try: search = driver.find_element(By.NAME, "text")
        except: search = driver.find_element(By.CSS_SELECTOR, "input[type='text']")
        
        query = f"schoolotzyv {school_name.replace('Государственное бюджетное общеобразовательное учреждение', '').replace('города Москвы', '').strip()}"
        search.send_keys(query)
        search.send_keys(Keys.RETURN)
        time.sleep(5)
        
        links = [l.get_attribute("href") for l in driver.find_elements(By.CSS_SELECTOR, "a[href*='schoolotzyv.ru']")]
        if not links:
            links = [l.get_attribute("href") for l in driver.find_elements(By.TAG_NAME, "a") if l.get_attribute("href") and "schoolotzyv.ru" in l.get_attribute("href")]
        
        url = next((l for l in links if 'schoolotzyv.ru/schools' in l), next((l for l in links if 'schoolotzyv.ru' in l and not any(x in l for x in ['search', 'static', 'yabs'])), None))
        if not url: return []
        
        driver.get(url)
        time.sleep(5)
        driver.execute_script("window.scrollTo(0, document.body.scrollHeight/2);")
        time.sleep(2)
        
        seen = set()
        for elem in driver.find_elements(By.CSS_SELECTOR, ".comment-item, .review-item, [class*='comment']"):
            txt = elem.text.strip()
            if len(txt) > 50 and not txt.startswith("Все отзывы") and re.search(r'#\d+|20\d{2}-\d{2}-\d{2}', txt):
                date = re.search(r'(\d{4}-\d{2}-\d{2})', txt)
                rating = "positive" if "Положительный" in txt else "negative" if "Отрицательный" in txt else "neutral"
                text_match = re.search(r'Это ложь\s+\d+\s*\n(.*?)(?:Ответить|$)', txt, re.DOTALL)
                text = text_match.group(1).strip() if text_match else txt
                
                key = (date.group(1) if date else "2024-01-01", text[:50])
                if key not in seen:
                    seen.add(key)
                    reviews.append({"date": key[0], "rating": rating, "text": text})
    except:
        pass
    finally:
        driver.quit()
    
    return reviews
