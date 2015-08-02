import base64, glob
import os.path
import subprocess

from xml.dom import minidom
from PIL import Image

root_path = os.path.dirname(__file__)

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

    tmp_svg_file = os.path.join(root_path, 'tmp.svg')
    with open(tmp_svg_file, 'w') as outSVG:
        tree.writexml(outSVG)

    filename, _ = os.path.splitext(os.path.basename(imagepath))
    outfilepath = os.path.join(root_path, filename + '.png')
    print "inkscape -z -e %s %s"  %(tmp_svg_file, outfilepath)
    subprocess.call("inkscape -z -e %s %s"  %(outfilepath, tmp_svg_file), shell=True)
    subprocess.call("mogrify -bordercolor black -trim  +repage -resize x65 -format png -quality 100 %s"  %(outfilepath), shell=True)

    os.rename (outfilepath, os.path.join(os.path.dirname(root_path),
                                         'logos',
                                         filename))

def main():
    for fil in glob.glob(os.path.join(root_path,'logos/*')):
        generate_svg(fil)

    os.remove(os.path.join(root_path, 'tmp.svg'))

main()
