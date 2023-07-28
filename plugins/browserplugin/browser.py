from flask import Flask, jsonify
from selenium import webdriver
from urllib.parse import urljoin
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import re
from bs4 import BeautifulSoup
from urllib.parse import unquote
import pytesseract
from pytesseract import Output
import cv2
import numpy as np
import undetected_chromedriver as uc
import sys


app = Flask(__name__)

@app.route('/')
def main():
    return jsonify(message='Hello, World!')

@app.route('/getcurrenturl/')
def getcurrenturl():
    try:
        message = "Current URL is: " + driver.current_url
    except:
        launchbrowser()
    return jsonify(message=message)

def launchbrowser():
    global driver
    driver = uc.Chrome()
    # return jsonify(messaZge='Success')

@app.route('/launchbrowser/')
def browserlaunch():
    launchbrowser()
    return jsonify(message="Success")


@app.route('/browserget/<path:value>/')
def browserget(value):
    decoded_url = unquote(value)
    url = 'https://' + decoded_url
    try:
        driver.get(url)
    except:
        launchbrowser()
        driver.get(url)
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message='Success')

@app.route('/browserclickcss/<selector>/')
def browserclickcss(selector):
    element = driver.find_element(By.CSS_SELECTOR, value=selector)
    # driver.execute_script('arguments[0].scrollIntoView(true);', element)
    element.click()
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message='Success') 

@app.route('/browserclickxpath///<selector>/')
def browserclickxpath(selector):
    element = driver.find_element(By.XPATH, value="//" + selector)
    # driver.execute_script('arguments[0].scrollIntoView(true);', element)
    element.click()
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message='Success') 

@app.route('/browsergettextcss/<selector>/')
def browsergettextcss(selector):
    res = driver.find_element(By.CSS_SELECTOR, value=selector).text
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message=res)

@app.route('/browsergettextxpath///<selector>/')
def browsergettextxpath(selector):
    res = driver.find_element(By.XPATH, value="//" + selector).text
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message=res) 

@app.route('/browsersendtextcss/<selector>/<text>/')
def browsersendtextcss(selector, text):
    element = driver.find_element(By.CSS_SELECTOR, value=selector)
    driver.execute_script('arguments[0].scrollIntoView(true);', element)
    element.send_keys(text)
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message='Success') 

@app.route('/browsersendtextxpath/<selector>///<text>/')
def browsersendtextxpath(selector, text):
    element = driver.find_element(By.XPATH, value="//" + selector)
    driver.execute_script('arguments[0].scrollIntoView(true);', element)
    element.send_keys(text)
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.XPATH, "//body")))
    return jsonify(message='Success') 

@app.route('/scrollup/')
def scrollup():
    driver.execute_script("window.scrollTo(0, window.scrollY - 500);")
    return jsonify(message='Success')

@app.route('/scrolldown/')
def scrolldown():
    driver.execute_script("window.scrollTo(0, window.scrollY + 500);")
    return jsonify(message='Success')

@app.route('/getbodytext/')
def getbodytext():
    text = driver.find_element(By.CSS_SELECTOR, value="body").text
    return jsonify(message=text)

@app.route('/gettagfromtext/<text>/')
def gettagfromtext(text):
    body = driver.find_element(By.CSS_SELECTOR, value="body").get_attribute("outerHTML")
    tag = re.findall(r'<([^>]+)>[^<]*' + re.escape(text), body, re.DOTALL)
    tagstring = ""
    for i in tag:
        tagstring += "[" + i + "]"
    return jsonify(message=tagstring)


@app.route('/getskeleton/')
def getskeleton():
    text = driver.find_element(By.CSS_SELECTOR, value="body").get_attribute("outerHTML")
    soup = BeautifulSoup(text, 'html.parser')
    result = ''

    for tag in soup.find_all():
        if tag.name in ['style', 'script', 'svg', 'path']:
            continue

        attrs = []
        for attr in ['name', 'type', 'value', 'alttext']: #'class', , 'href'
            if attr in tag.attrs:
                attrs.append(f'{attr}="{tag[attr]}"')
        
        tag_string = f'<{tag.name} {" ".join(attrs)}>'

        if tag.string:
            tag_string += tag.string.rstrip()
    
        result += tag_string

    result = result.replace("<div >", "") 
    result = result.replace("<span >", "") 
    result = result.replace("<body >", "")
    result = result.replace(" >", "")

    return jsonify(message=result) 


@app.route('/gettextcoordinates/')
def gettextcoordinates():
    screenshot_data = driver.get_screenshot_as_png()

    # Convert the PNG data to an image array
    screenshot_array = np.frombuffer(screenshot_data, np.uint8)
    img = cv2.imdecode(screenshot_array, cv2.IMREAD_UNCHANGED)
    # Convert image to grey scale
    gray_image = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    thresh_img = cv2.threshold(gray_image, 0, 255, cv2.THRESH_BINARY | cv2.THRESH_OTSU)[1]
    custom_config = r'--oem 3 --psm 3'
    ocr_output_details = pytesseract.image_to_data(thresh_img, output_type=Output.DICT, config=custom_config, lang='eng')

    output = ""
    for i in range(len(ocr_output_details["text"])):
        if not ocr_output_details["text"][i]:
            continue
        #calculate mid
        middle_x = ocr_output_details["left"][i] + ocr_output_details["width"][i] / 2
        middle_y = ocr_output_details["top"][i] + ocr_output_details["height"][i] / 2
        output += "[text: " + ocr_output_details["text"][i] + ", mid: " + str(middle_x) + ", " + str(middle_y) + "]"
    
    print(ocr_output_details)

    return jsonify(message=output)


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8210, threaded=True)
