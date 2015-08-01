
from xml.dom import minidom
from PIL import Image
import base64
import imghdr

def generate_svg(image):
    im = Image.open(image)
    format = im.format.lower()
    width, height = im.size

    tree = minidom.parse('t.svg')
    imagetag = tree.getElementsByTagName('image')[0]

    imagetag.setAttribute('width', str(width))
    imagetag.setAttribute('height', str(height))

    with open(image, "rb") as imagefile:
        base64data = base64.b64encode(imagefile.read())
    imagetag.setAttribute('xlink:href', 'data:image/' + format + ';base64,' + base64data)

    with open('out.svg', 'w') as outSVG:
        tree.writexml(outSVG)

generate_svg('logos/fpty_khume_logo.gif')
