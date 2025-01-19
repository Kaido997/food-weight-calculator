# 🍽️ Food Weight Calculator

[![Website](https://img.shields.io/badge/Live_Site-Online-brightgreen)](https://calculator.foodweight.online/)
[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://golang.org/)

Food Weight Calculator is a simple web-based application that estimates the cooked weight of food based on its raw weight. It helps users make informed decisions about portion sizes and nutrition.

🔗 **Live Website:** [calculator.foodweight.online](https://calculator.foodweight.online/)
📂 **GitHub Repository:** [Kaido997/food-weight-calculator](https://github.com/Kaido997/food-weight-calculator)

## 🛠️ Features

- Convert raw food weight to cooked weight instantly.
- Support for various food types with different cooking weight loss factors.
- Responsive and lightweight web interface.
- Built using Go (without frameworks) for high performance and simplicity.

## 🏗️ Project Structure

```
food-weight-calculator/
│-- api/          # Handles API requests for weight calculations
│-- internal/     # Internal logic and database management
│-- services/     # Core business logic and calculations
│-- web/          # HTML, CSS, and JS frontend files
│-- main.go       # Entry point of the application
│-- go.mod        # Go dependencies
│-- Dockerfile    # Docker setup for deployment
│-- README.md     # Project documentation
```

## 🚀 Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Kaido997/food-weight-calculator.git
   cd food-weight-calculator
   ```

2. **Run the application:**
   ```bash
   go run main.go
   ```
   or
   ```bash
   air
   ```

3. **Access the website:**
   Open `http://localhost:8080` in your web browser.

## 🌐 Usage

1. Enter the raw food weight (grams).
2. Choose the food type (e.g., chicken, beef, fish, vegetables).
3. View the estimated cooked weight instantly.

## 📝 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on GitHub.

## 📧 Contact

For questions or suggestions, reach out via:

- GitHub: [Kaido997](https://github.com/Kaido997)
