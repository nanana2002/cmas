from PIL import Image
import os
def split_image():
    """
    用户输入图片路径和分割张数，将图片从上到下等分，保存到同名文件夹。
    """
    img_path = input("请输入图片路径：").strip()
    num_slices = int(input("请输入要分割的张数：").strip())

    if not os.path.isfile(img_path):
        print("图片路径不存在！")
        return

    base_name = os.path.basename(img_path)
    name, ext = os.path.splitext(base_name)
    folder = os.path.join(os.path.dirname(img_path), name)
    os.makedirs(folder, exist_ok=True)

    img = Image.open(img_path)
    width, height = img.size
    slice_height = height // num_slices

    for i in range(num_slices):
        top = i * slice_height
        bottom = (i + 1) * slice_height if i < num_slices - 1 else height
        box = (0, top, width, bottom)
        slice_img = img.crop(box)
        out_name = f"{name}_{i+1}{ext}"
        out_path = os.path.join(folder, out_name)
        slice_img.save(out_path)
        print(f"保存: {out_path}")

    print(f"分割完成，图片保存在: {folder}")


if __name__ == '__main__':
    split_image()