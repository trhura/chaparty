import base64, glob
import os.path
import subprocess
import random
import itertools
import shutil

from xml.dom import minidom
from PIL import Image

root_path = os.path.dirname(os.path.abspath(__file__))

light_colors = [
    "#ffcdd2",
    #"#ede7f6",
    #"#b9f6ca",
    "#ccff90",
    #"#ff9e80",
    "#ffff8d",
    "#f5f5f5",
]

dark_colors = [
    "#ff6600",
]

def generate_svg(imagepath):
    img = Image.open(imagepath)
    fmt = img.format.lower()
    width, height = img.size

    template_path = os.path.join(root_path, 'template.svg')
    tree = minidom.parse(template_path)
    imagetag = tree.getElementsByTagName('image')[0]

    aheight = 70
    awidth = aheight * width / height
    imagetag.setAttribute('width', str(awidth))
    imagetag.setAttribute('height', str(aheight))

    with open(imagepath, "rb") as imagefile:
        base64data = base64.b64encode(imagefile.read())
    imagetag.setAttribute('xlink:href', 'data:image/' + fmt + ';base64,' + base64data)

    filename, _ = os.path.splitext(os.path.basename(imagepath))
    outputdir = os.path.join(os.path.dirname(root_path), 'logos', filename)
    if os.path.exists(outputdir): shutil.rmtree(outputdir)
    os.mkdir(outputdir)

    for i, colorpair in enumerate(itertools.product(light_colors, dark_colors)):
        bg, fg = colorpair
        recttag = tree.getElementsByTagName('rect')[0]
        style = recttag.getAttribute('style')
        style = 'fill:' + bg + style[12:]
        recttag.setAttribute('style', style)

        pathtag = tree.getElementsByTagName('path')[0]
        style = recttag.getAttribute('style')
        style = 'fill:' + fg + style[12:] + ';stroke-opacity:1'
        pathtag.setAttribute('style', style)

        tmp_svg_file = os.path.join(root_path, 'tmp.svg')
        with open(tmp_svg_file, 'w') as outSVG:
            tree.writexml(outSVG)

        tmpfilepath = os.path.join(root_path, filename + '.png')
        subprocess.call("inkscape -z -e %s %s"  %(tmpfilepath, tmp_svg_file), shell=True)
        subprocess.call("""
        mogrify -bordercolor black -trim  +repage -resize x65 -format png -quality 100 %s
        """  %(tmpfilepath), shell=True)

        os.rename (tmpfilepath, os.path.join(outputdir, str(i)))

def main():
    for fil in glob.glob(os.path.join(root_path,'logos/*')):
        generate_svg(fil)

    os.remove(os.path.join(root_path, 'tmp.svg'))

main()
