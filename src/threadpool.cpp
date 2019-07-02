/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

Author: nanjj

*/

#include <iostream>
#include <chrono>
#include <string>
#include <mutex>

#include "threadpool.hpp"

using namespace std;
const int THREAD_COUNT = 10;
mutex g_display_mutex;

class TestThreadpool
{
private:
    ThreadPool<THREAD_COUNT> pool;
public:
    void ExecuteAsync(int i, char *s) {
        s = strdup(s);
        auto job = [this,i,s]{
                       this->Execute(i, s);
                       free(s);
                   };
        pool.AddJob(job);
    };
    void Execute(int i, char *s) {
        g_display_mutex.lock();
        cout << s << "-> @" << this_thread::get_id() << "@" <<i << endl;
        g_display_mutex.unlock();
        this_thread::sleep_for( chrono::seconds((i+1)/(i+1)) );
        g_display_mutex.lock();
        cout << s << "<-@" << this_thread::get_id() << "@" <<i << endl;
        g_display_mutex.unlock();
    }
};

int main() {
    int JOB_COUNT = 100;
    TestThreadpool t;
    for( int i = 0; i < JOB_COUNT; ++i ) {
        char s[10];
        sprintf(s, "%d", i);
        t.ExecuteAsync(i, s);
    }
    cout << "Expected runtime: " << JOB_COUNT/THREAD_COUNT<<" seconds." << endl;
}
