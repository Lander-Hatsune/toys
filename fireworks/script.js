const PI = 355 / 113;

const COLOR = {
    BG: 'rgba(10, 10, 30, 0.3)',
    HALO: 'rgba(255, 255, 255, 0.1)',
}

const SHELL = {
    R: 5, // size
    HALO_R: 10, // halo size
    V: 2, // initial velocity
    A: 2, // acceleration
    HUE_L: 30, // minimum hue
    HUE_R: 150, // maximum hue

    N: 70, // split into how many stars (mean)
    N_STD: 5, // split into how many stars (std)
}

const STAR = {
    R: 3, // size
    HUE_STD: 10, // hue std (mean val is 'huebase')
    SATUR_STD: 30, // saturation std (mean val is 100)
    
    V: 50, // initial velocity
    V_STD: 20, // velocity std

    RESIS: 0.01, // air resistance
    GRAV_A: 0.1, // gravity accel
    CURV_A: 0.1, // curving accel
    
    BLINK_THR: 0.8, // blink probability, 1 is forever visible
    TTL: 100, // time to live
}

function clip(x, minn, maxx) {
    return min(max(x, minn), maxx);
}

class Star {
    constructor(origin, huebase) {
        this.r = STAR.R;
        this.pos = createVector(origin.x, origin.y);
        this.pos_ = createVector(origin.x, origin.y); // draw trajectory for smooth
        this.hue = clip(round(randomGaussian(huebase, STAR.HUE_STD)), 0, 360);
        this.satur = clip(round(randomGaussian(100, STAR.SATUR_STD)), 0, 100);

        let vm = randomGaussian(STAR.V, STAR.V_STD); // velocity magnitude
        let vd = random(0, 2 * PI); // velocity direction
        this.v = createVector(vm * cos(vd), vm * -sin(vd));

        this.grav_a = createVector(0, STAR.GRAV_A);
        let curv_d = random(-PI / 2, PI / 2) + vd;
        this.curv_a = createVector(STAR.CURV_A * cos(curv_d),
                                   STAR.CURV_A * -sin(curv_d));
        this.visible = true;
        this.ttl = STAR.TTL;
        this.color = color(this.hue, 100, 100);
    }

    update() {
        this.ttl = max(0, this.ttl - 1);
        if (this.ttl > 0) {
            let a = p5.Vector.sub( // gravity + curving + resistence
                p5.Vector.add(this.grav_a, this.curv_a),
                p5.Vector.mult(this.v, STAR.RESIS * this.v.mag()))
            this.v.add(a);
            this.pos_ = createVector(this.pos.x, this.pos.y);
            this.pos.add(this.v);
            let thr = this.visible ? STAR.BLINK_THR : 1 - STAR.BLINK_THR;
            let blink = (random(0, 1) > thr);
            if (blink) {
                this.visible = !this.visible;
                if (this.visible) {
                    this.r = 2 * STAR.R;
                }
            } else {
                this.r = max(this.r - 1, STAR.R);
            }
            let tr = this.ttl / STAR.TTL; // ttl ratio
            this.color = color(this.hue, this.satur, sq(tr) * 100);
        }
    }

    draw() {
        if (this.ttl > 0 && this.visible) {
            stroke(this.color);
            strokeWeight(this.r);
            line(this.pos_.x, this.pos_.y, this.pos.x, this.pos.y);
        }
    }
}

class Shell {
    constructor(targx, targy) {
        this.targ = createVector(targx, targy);
        this.pos = createVector(width / 2, height - 5);
        this.pos_ = createVector(this.pos.x, this.pos.y);
        this.start = createVector(this.pos.x, this.pos.y);
        this.huebase = random(SHELL.HUE_L, SHELL.HUE_R);
        this.color = color(round(this.huebase), 70, 70);
        let d = atan2(this.targ.y - this.pos.y,
                      this.targ.x - this.pos.x); // direction of v & a
        this.v = createVector(SHELL.V * cos(d), SHELL.V * sin(d));
        this.a = createVector(SHELL.A * cos(d), SHELL.V * sin(d));
        this.stars = [];
        this.exploded = false;
        this.over = false;
    }
    explode() {
        this.exploded = true;
        for (let i = 0; i < randomGaussian(SHELL.N, SHELL.N_STD); i++) {
            this.stars.push(new Star(this.targ, this.huebase));
        }
    }
    draw() {
        if (!this.exploded) {
            stroke(COLOR.HALO);
            strokeWeight(SHELL.HALO_R);
            line(this.pos_.x, this.pos_.y, this.pos.x, this.pos.y);
            stroke(this.color);
            strokeWeight(SHELL.R);
            line(this.pos_.x, this.pos_.y, this.pos.x, this.pos.y);
        } else {
            this.stars.forEach(star => {
                star.draw();
            });
        }
    }
    update() {
        if (!this.exploded) {
            this.v.add(this.a);
            this.pos_ = createVector(this.pos.x, this.pos.y);
            this.pos.add(this.v);
            if (this.start.dist(this.pos) >= this.start.dist(this.targ)) {
                this.explode();
            }
        } else {
            this.stars.forEach((star, idx, arr) => {
                star.update();
                if (star.ttl <= 0) arr.splice(idx, 1);
                if (this.stars.length == 0) this.over = true;
            });
        }
    }
}

let shells;
function mousePressed() {
    shells.push(new Shell(mouseX, mouseY));
}
function mouseDragged() {
    shells.push(new Shell(mouseX, mouseY));
}

function setup() {
    colorMode(HSB, 360, 100, 100);
    createCanvas((displayWidth - 100) / 2, displayHeight - 200);
    shells = [];
}

function draw() {
    background(COLOR.BG);
    shells.forEach((shell, idx, arr) => {
        shell.update();
        shell.draw();
        if (shell.over) arr.splice(idx, 1);
    });
}
