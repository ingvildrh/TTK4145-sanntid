# Python 3.3.3 and 2.7.6
# python helloworld_python.py

from threading import Thread

j = 0

def thread_1_function():
    global j
    for i in range(0,100000):
        j+=1

def thread_2_function():
    global j
    for i in range(-100000,0):
        j-=1

def main():
    thread_1 = Thread(target = thread_1_function, args = (),)
    thread_1.start()
    thread_2 = Thread(target = thread_2_function, args = (),)
    thread_2.start()

    thread_1.join()
    thread_2.join()
    print(j)

main()
