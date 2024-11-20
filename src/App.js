import React from 'react';
import {BrowserRouter as Router, Routes, Route} from 'react-router-dom';
import './App.css';
import Login from './features/Login/Login';
import SignUp from './features/SignUp/SignUp';
import Dashboard from './features/Dashboard/Dashboard';
import ProtectedRoute from './components/ProtectedRoute';
import Profile from './features/Profile/Profile';
import Settings from './features/Settings/Settings';
import AddTransactions from './features/AddTransactions/AddTransactions';
import CategoryList from './features/CategoryList/CategoryList';
import IconCatalogList from './features/IconCatalogList/IconCatalogList';
import CreateCategory from './features/CreateCategory/CreateCategory';
import AddTransactionIncome from './features/AddTransactionIncome/AddTransactionIncome';

function App() {
    const handleLogin = () => {
    };
    
    const handleSignUp = () => {
    };
    
    return (
        <div className="App">
            <Routes>
                <Route path="/signup" element={<SignUp onSignUp={handleSignUp}/>}/>
                <Route path="/login" element={<Login onLogin={handleLogin}/>}/>
                <Route path="/" element={<Login onLogin={handleLogin}/>}/>
                <Route path="/dashboard" element={
                    <ProtectedRoute>
                        <Dashboard/>
                    </ProtectedRoute>
                }/>
                <Route path="/profile" element={
                    <ProtectedRoute>
                        <Profile/>
                    </ProtectedRoute>
                }/>
                <Route path={"/settings"} element={
                    <ProtectedRoute>
                        <Settings/>
                    </ProtectedRoute>
                }/>
                <Route path={"/add-transactions"} element={
                    <ProtectedRoute>
                        <AddTransactions/>
                    </ProtectedRoute>
                }/>
                <Route path={"/list-categories"} element={
                    <ProtectedRoute>
                        <CategoryList/>
                    </ProtectedRoute>
                }/>
                <Route path={"/list-icon-catalog"} element={
                    <ProtectedRoute>
                        <IconCatalogList/>
                    </ProtectedRoute>
                }/>
                <Route path="/create-category" element={
                    <ProtectedRoute>
                        <CreateCategory/>
                    </ProtectedRoute>
                }/>
                <Route path="/add-transaction-income" element={
                    <ProtectedRoute>
                        <AddTransactionIncome />
                    </ProtectedRoute>
                } />
            </Routes>
        </div>
    );
}

export default function AppWrapper() {
    return (
        <Router>
            <App/>
        </Router>
    );
}