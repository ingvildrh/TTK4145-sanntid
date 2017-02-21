#include <pthread.h>
#include <stdio.h>

int j;
pthread_mutex_t count_mutex;


void* thread_1_function() {
	for (int i = 0; i < 1000001; i++) {
		pthread_mutex_lock(&count_mutex);
		j++;
		pthread_mutex_unlock(&count_mutex);
	}
  return NULL;
}

void* thread_2_function() {
	for (int i = 0; i < 1000000; i++) {
		pthread_mutex_lock(&count_mutex);
		j--;
		pthread_mutex_unlock(&count_mutex);
	}
  return NULL;
}

int main(){
    j = 0;
    pthread_t thread_1;
    pthread_create(&thread_1, NULL, thread_1_function, NULL);
    pthread_t thread_2;
    pthread_create(&thread_2, NULL, thread_2_function, NULL);

    pthread_join(thread_1, NULL);
    pthread_join(thread_2, NULL);

    printf("\tawesomesauce: %d\n", j);

    return 0;

}
