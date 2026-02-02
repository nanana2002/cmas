import os
import fitz  # PyMuPDF

def pdf_first_page_to_png(folder_path):
    photo_folder = os.path.join(folder_path, "photo")
    if not os.path.exists(photo_folder):
        os.makedirs(photo_folder)

    for filename in os.listdir(folder_path):
        if filename.lower().endswith(".pdf"):
            pdf_path = os.path.join(folder_path, filename)
            doc = fitz.open(pdf_path)
            found = False
            for page_num in range(doc.page_count):
                page = doc.load_page(page_num)
                text = page.get_text().lower()
                if "abstract" in text or "摘要" in text:
                    pix = page.get_pixmap()
                    png_name = os.path.splitext(filename)[0] + ".png"
                    png_path = os.path.join(photo_folder, png_name)
                    pix.save(png_path)
                    found = True
                    break
            doc.close()

if __name__ == "__main__":
    folder = input("请输入PDF所在文件夹路径：")
    pdf_first_page_to_png(folder)
    print("转换完成！")