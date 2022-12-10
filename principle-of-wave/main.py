import numpy as np
import matplotlib.pyplot as plt

def circ(theta, t, x_shift=0, y_shift=0):
    return (theta + np.cos(theta + t) + x_shift,
            np.sin(theta + t) + y_shift)

DELTA_T = 0.05 # 20fps

center_circle = circ(0, np.linspace(0, 2 * np.pi))

t = 0
while t < 10:
    plt.clf()
    t += DELTA_T
    centers = np.arange(-2.5 * np.pi, 2.5 * np.pi, np.pi * DELTA_T)
    plt.axis('equal')
    plt.xlim(-3 * np.pi, 3 * np.pi)
    plt.scatter(*circ(centers, t * np.pi), c='b', alpha=0.4)
    plt.scatter(*circ(centers, t * np.pi, 0.3, 0.5), c='b', alpha=0.2)
    plt.scatter(*circ(centers, t * np.pi, 0.6, 1.0), c='b', alpha=0.1)
    plt.plot(*center_circle, c='r')
    plt.scatter(*circ(0, t * np.pi), c='r')
    plt.pause(DELTA_T)

