from locust import HttpUser, task, between
import random

class WeatherTweetUser(HttpUser):
    wait_time = between(1, 3)
    
    municipalities = ["guatemala"]
    
    # Todos los climas posibles
    weathers = ["sunny", "cloudy", "rainy", "foggy"]
    
    @task
    def send_weather_tweet(self):
        tweet = {
            "municipality": random.choice(self.municipalities),
            "temperature": random.randint(15, 35),  # Temperatura entre 15-35°C
            "humidity": random.randint(40, 90),     # Humedad entre 40-90%
            "weather": random.choice(self.weathers)
        }
        
        response = self.client.post("/tweet", json=tweet)
        
        if response.status_code != 200:
            print(f"Error: {response.status_code} - {response.text}")