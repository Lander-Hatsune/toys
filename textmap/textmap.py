import argparse
import numpy as np
from PIL import Image

parser = argparse.ArgumentParser()
parser.add_argument(
    "bg", type=str,
    help="background character")
parser.add_argument(
    "fg", type=str,
    help="foreground character")
parser.add_argument(
    "charimg", type=str,
    help="character image")
parser.add_argument(
    "--width", type=int,
    default=12,
    help="map width")
args = parser.parse_args()

img = Image.open(args.charimg).convert("1")
img = img.resize((args.width, img.height * args.width // img.width))
imgarr = np.array(img)

print("resized:", imgarr.shape)

textmap = [[args.bg if pix else args.fg for pix in line] for line in imgarr]

for line in textmap:
    for ch in line:
        print(ch, end="")
    print()
    




