from locust import FastHttpUser, task, between
import random


class ProductSearchUser(FastHttpUser):
    """
    Load test for the product search service.
    Searches for common terms from our real product catalog.
    """

    # Minimal wait time to maximize load
    # wait_time = between(0.1, 0.5)

    # Search terms based on real DummyJSON product names and categories
    search_terms = [
        "beauty",
        "fragrances",
        "furniture",
        "groceries",
        "chicken",
        "apple",
        "mascara",
        "lipstick",
        "calvin",
        "chanel",
        "gucci",
        "bed",
        "sofa",
        "table",
        "steak",
        "cat",
        "dog",
        "powder",
        "mirror",
        "cherry",
    ]

    @task(10)
    def search_products(self):
        """Search for products using rotating search terms"""
        term = random.choice(self.search_terms)
        self.client.get(f"/products/search?q={term}", name="/products/search")

    @task(1)
    def health_check(self):
        """Periodic health check"""
        self.client.get("/health")